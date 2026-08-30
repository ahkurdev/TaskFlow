package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"taskflow/internal/model"
	"taskflow/internal/repository"
)

type ListService struct {
	boards      *repository.BoardRepository
	lists       *repository.ListRepository
	pool        *pgxpool.Pool
	broadcaster Broadcaster
}

func NewListService(boards *repository.BoardRepository, lists *repository.ListRepository, pool *pgxpool.Pool, broadcaster Broadcaster) *ListService {
	return &ListService{boards: boards, lists: lists, pool: pool, broadcaster: withDefaultBroadcaster(broadcaster)}
}

func (s *ListService) Create(ctx context.Context, userID, boardID int64, name string) (*model.List, error) {
	if _, err := s.verifyBoardOwnership(ctx, userID, boardID); err != nil {
		return nil, err
	}

	existing, err := s.lists.ListByBoard(ctx, boardID)
	if err != nil {
		return nil, err
	}

	list, err := s.lists.Create(ctx, boardID, name, len(existing))
	if err != nil {
		return nil, err
	}

	s.broadcaster.Broadcast(boardID, EventBoardUpdated)
	return list, nil
}

func (s *ListService) Update(ctx context.Context, userID, boardID, listID int64, name string) (*model.List, error) {
	if _, err := s.verifyBoardOwnership(ctx, userID, boardID); err != nil {
		return nil, err
	}
	if _, err := s.verifyListInBoard(ctx, boardID, listID); err != nil {
		return nil, err
	}

	updated, err := s.lists.Update(ctx, listID, name)
	if err != nil {
		return nil, err
	}

	s.broadcaster.Broadcast(boardID, EventBoardUpdated)
	return updated, nil
}

// Reorder moves a list to newPosition (0-indexed) within its board and
// resequences all sibling lists accordingly.
func (s *ListService) Reorder(ctx context.Context, userID, boardID, listID int64, newPosition int) error {
	if _, err := s.verifyBoardOwnership(ctx, userID, boardID); err != nil {
		return err
	}
	if _, err := s.verifyListInBoard(ctx, boardID, listID); err != nil {
		return err
	}

	lists, err := s.lists.ListByBoard(ctx, boardID)
	if err != nil {
		return err
	}

	ordered := reorderByID(lists, listID, newPosition, func(l model.List) int64 { return l.ID })

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	txLists := repository.NewListRepository(tx)
	for i, l := range ordered {
		if err := txLists.UpdatePosition(ctx, l.ID, i); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	s.broadcaster.Broadcast(boardID, EventBoardUpdated)
	return nil
}

func (s *ListService) Delete(ctx context.Context, userID, boardID, listID int64) error {
	if _, err := s.verifyBoardOwnership(ctx, userID, boardID); err != nil {
		return err
	}
	if _, err := s.verifyListInBoard(ctx, boardID, listID); err != nil {
		return err
	}

	if err := s.lists.Delete(ctx, listID); err != nil {
		return err
	}

	remaining, err := s.lists.ListByBoard(ctx, boardID)
	if err != nil {
		return err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	txLists := repository.NewListRepository(tx)
	for i, l := range remaining {
		if err := txLists.UpdatePosition(ctx, l.ID, i); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	s.broadcaster.Broadcast(boardID, EventBoardUpdated)
	return nil
}

func (s *ListService) verifyBoardOwnership(ctx context.Context, userID, boardID int64) (*model.Board, error) {
	board, err := s.boards.GetByID(ctx, boardID)
	if err != nil {
		if errors.Is(err, repository.ErrBoardNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get board: %w", err)
	}
	if board.UserID != userID {
		return nil, ErrNotFound
	}
	return board, nil
}

func (s *ListService) verifyListInBoard(ctx context.Context, boardID, listID int64) (*model.List, error) {
	list, err := s.lists.GetByID(ctx, listID)
	if err != nil {
		if errors.Is(err, repository.ErrListNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get list: %w", err)
	}
	if list.BoardID != boardID {
		return nil, ErrNotFound
	}
	return list, nil
}

// reorderByID removes the item identified by id from items, clamps
// newPosition to a valid index, and re-inserts it there, returning the
// full resulting order.
func reorderByID[T any](items []T, id int64, newPosition int, idOf func(T) int64) []T {
	filtered := make([]T, 0, len(items))
	var moved T
	for _, item := range items {
		if idOf(item) == id {
			moved = item
			continue
		}
		filtered = append(filtered, item)
	}

	if newPosition < 0 {
		newPosition = 0
	}
	if newPosition > len(filtered) {
		newPosition = len(filtered)
	}

	result := make([]T, 0, len(items))
	result = append(result, filtered[:newPosition]...)
	result = append(result, moved)
	result = append(result, filtered[newPosition:]...)

	return result
}

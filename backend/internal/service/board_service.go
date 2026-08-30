package service

import (
	"context"
	"errors"
	"fmt"

	"taskflow/internal/model"
	"taskflow/internal/repository"
)

type BoardService struct {
	boards      *repository.BoardRepository
	lists       *repository.ListRepository
	cards       *repository.CardRepository
	broadcaster Broadcaster
}

func NewBoardService(boards *repository.BoardRepository, lists *repository.ListRepository, cards *repository.CardRepository, broadcaster Broadcaster) *BoardService {
	return &BoardService{boards: boards, lists: lists, cards: cards, broadcaster: withDefaultBroadcaster(broadcaster)}
}

func (s *BoardService) Create(ctx context.Context, userID int64, name, description string) (*model.Board, error) {
	return s.boards.Create(ctx, userID, name, description)
}

func (s *BoardService) ListForUser(ctx context.Context, userID int64) ([]model.Board, error) {
	return s.boards.ListByUser(ctx, userID)
}

// GetDetail returns a board with its lists and cards, provided it belongs
// to the requesting user.
func (s *BoardService) GetDetail(ctx context.Context, userID, boardID int64) (*model.BoardDetail, error) {
	board, err := s.verifyOwnership(ctx, userID, boardID)
	if err != nil {
		return nil, err
	}

	lists, err := s.lists.ListByBoard(ctx, boardID)
	if err != nil {
		return nil, err
	}

	detail := &model.BoardDetail{Board: *board, Lists: make([]model.ListWithCards, 0, len(lists))}
	for _, l := range lists {
		cards, err := s.cards.ListByList(ctx, l.ID)
		if err != nil {
			return nil, err
		}
		detail.Lists = append(detail.Lists, model.ListWithCards{List: l, Cards: cards})
	}

	return detail, nil
}

func (s *BoardService) Update(ctx context.Context, userID, boardID int64, name, description string) (*model.Board, error) {
	if _, err := s.verifyOwnership(ctx, userID, boardID); err != nil {
		return nil, err
	}

	board, err := s.boards.Update(ctx, boardID, name, description)
	if err != nil {
		return nil, err
	}

	s.broadcaster.Broadcast(boardID, EventBoardUpdated)
	return board, nil
}

func (s *BoardService) Delete(ctx context.Context, userID, boardID int64) error {
	if _, err := s.verifyOwnership(ctx, userID, boardID); err != nil {
		return err
	}

	if err := s.boards.Delete(ctx, boardID); err != nil {
		return err
	}

	s.broadcaster.Broadcast(boardID, EventBoardUpdated)
	return nil
}

// VerifyOwnership confirms boardID exists and belongs to userID, without
// fetching the full board detail. Used by the WebSocket handler to
// authorize a subscription.
func (s *BoardService) VerifyOwnership(ctx context.Context, userID, boardID int64) error {
	_, err := s.verifyOwnership(ctx, userID, boardID)
	return err
}

// verifyOwnership fetches a board and ensures it belongs to userID.
func (s *BoardService) verifyOwnership(ctx context.Context, userID, boardID int64) (*model.Board, error) {
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

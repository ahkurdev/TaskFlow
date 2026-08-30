package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"taskflow/internal/model"
	"taskflow/internal/repository"
)

type CardService struct {
	boards      *repository.BoardRepository
	lists       *repository.ListRepository
	cards       *repository.CardRepository
	pool        *pgxpool.Pool
	broadcaster Broadcaster
}

func NewCardService(boards *repository.BoardRepository, lists *repository.ListRepository, cards *repository.CardRepository, pool *pgxpool.Pool, broadcaster Broadcaster) *CardService {
	return &CardService{boards: boards, lists: lists, cards: cards, pool: pool, broadcaster: withDefaultBroadcaster(broadcaster)}
}

func (s *CardService) Create(ctx context.Context, userID, boardID, listID int64, title, description string) (*model.Card, error) {
	if _, err := s.verifyListInBoard(ctx, userID, boardID, listID); err != nil {
		return nil, err
	}

	existing, err := s.cards.ListByList(ctx, listID)
	if err != nil {
		return nil, err
	}

	card, err := s.cards.Create(ctx, listID, title, description, len(existing))
	if err != nil {
		return nil, err
	}

	s.broadcaster.Broadcast(boardID, EventBoardUpdated)
	return card, nil
}

func (s *CardService) Update(ctx context.Context, userID, boardID, listID, cardID int64, title, description string) (*model.Card, error) {
	if _, err := s.verifyListInBoard(ctx, userID, boardID, listID); err != nil {
		return nil, err
	}
	if _, err := s.verifyCardInList(ctx, listID, cardID); err != nil {
		return nil, err
	}

	card, err := s.cards.Update(ctx, cardID, title, description)
	if err != nil {
		return nil, err
	}

	s.broadcaster.Broadcast(boardID, EventBoardUpdated)
	return card, nil
}

// Move relocates a card to targetListID at newPosition (0-indexed), which
// may be the same list (simple reorder) or a different one (cross-list move).
// Both the source and target lists must belong to boardID.
func (s *CardService) Move(ctx context.Context, userID, boardID, sourceListID, cardID, targetListID int64, newPosition int) (*model.Card, error) {
	if _, err := s.verifyListInBoard(ctx, userID, boardID, sourceListID); err != nil {
		return nil, err
	}
	if _, err := s.verifyListInBoard(ctx, userID, boardID, targetListID); err != nil {
		return nil, err
	}
	card, err := s.verifyCardInList(ctx, sourceListID, cardID)
	if err != nil {
		return nil, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	txCards := repository.NewCardRepository(tx)

	if sourceListID == targetListID {
		cards, err := s.cards.ListByList(ctx, sourceListID)
		if err != nil {
			return nil, err
		}
		ordered := reorderByID(cards, cardID, newPosition, func(c model.Card) int64 { return c.ID })
		for i, c := range ordered {
			if err := txCards.UpdatePositionAndList(ctx, c.ID, sourceListID, i); err != nil {
				return nil, err
			}
		}
	} else {
		sourceCards, err := s.cards.ListByList(ctx, sourceListID)
		if err != nil {
			return nil, err
		}
		targetCards, err := s.cards.ListByList(ctx, targetListID)
		if err != nil {
			return nil, err
		}

		remainingSource := make([]model.Card, 0, len(sourceCards))
		for _, c := range sourceCards {
			if c.ID != cardID {
				remainingSource = append(remainingSource, c)
			}
		}

		if newPosition < 0 {
			newPosition = 0
		}
		if newPosition > len(targetCards) {
			newPosition = len(targetCards)
		}
		newTargetOrder := make([]model.Card, 0, len(targetCards)+1)
		newTargetOrder = append(newTargetOrder, targetCards[:newPosition]...)
		newTargetOrder = append(newTargetOrder, *card)
		newTargetOrder = append(newTargetOrder, targetCards[newPosition:]...)

		for i, c := range remainingSource {
			if err := txCards.UpdatePositionAndList(ctx, c.ID, sourceListID, i); err != nil {
				return nil, err
			}
		}
		for i, c := range newTargetOrder {
			if err := txCards.UpdatePositionAndList(ctx, c.ID, targetListID, i); err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	s.broadcaster.Broadcast(boardID, EventBoardUpdated)
	return s.cards.GetByID(ctx, cardID)
}

func (s *CardService) Delete(ctx context.Context, userID, boardID, listID, cardID int64) error {
	if _, err := s.verifyListInBoard(ctx, userID, boardID, listID); err != nil {
		return err
	}
	if _, err := s.verifyCardInList(ctx, listID, cardID); err != nil {
		return err
	}

	if err := s.cards.Delete(ctx, cardID); err != nil {
		return err
	}

	remaining, err := s.cards.ListByList(ctx, listID)
	if err != nil {
		return err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	txCards := repository.NewCardRepository(tx)
	for i, c := range remaining {
		if err := txCards.UpdatePositionAndList(ctx, c.ID, listID, i); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	s.broadcaster.Broadcast(boardID, EventBoardUpdated)
	return nil
}

func (s *CardService) verifyListInBoard(ctx context.Context, userID, boardID, listID int64) (*model.List, error) {
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

func (s *CardService) verifyCardInList(ctx context.Context, listID, cardID int64) (*model.Card, error) {
	card, err := s.cards.GetByID(ctx, cardID)
	if err != nil {
		if errors.Is(err, repository.ErrCardNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get card: %w", err)
	}
	if card.ListID != listID {
		return nil, ErrNotFound
	}
	return card, nil
}

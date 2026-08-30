package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"taskflow/internal/model"
)

var ErrCardNotFound = errors.New("card not found")

type CardRepository struct {
	db Querier
}

func NewCardRepository(db Querier) *CardRepository {
	return &CardRepository{db: db}
}

func (r *CardRepository) Create(ctx context.Context, listID int64, title, description string, position int) (*model.Card, error) {
	const query = `
		INSERT INTO cards (list_id, title, description, position)
		VALUES ($1, $2, $3, $4)
		RETURNING id, list_id, title, description, position, created_at, updated_at
	`

	var c model.Card
	err := r.db.QueryRow(ctx, query, listID, title, description, position).Scan(
		&c.ID, &c.ListID, &c.Title, &c.Description, &c.Position, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert card: %w", err)
	}

	return &c, nil
}

func (r *CardRepository) GetByID(ctx context.Context, id int64) (*model.Card, error) {
	const query = `
		SELECT id, list_id, title, description, position, created_at, updated_at
		FROM cards
		WHERE id = $1
	`

	var c model.Card
	err := r.db.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.ListID, &c.Title, &c.Description, &c.Position, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCardNotFound
		}
		return nil, fmt.Errorf("get card: %w", err)
	}

	return &c, nil
}

func (r *CardRepository) ListByList(ctx context.Context, listID int64) ([]model.Card, error) {
	const query = `
		SELECT id, list_id, title, description, position, created_at, updated_at
		FROM cards
		WHERE list_id = $1
		ORDER BY position ASC
	`

	rows, err := r.db.Query(ctx, query, listID)
	if err != nil {
		return nil, fmt.Errorf("list cards: %w", err)
	}
	defer rows.Close()

	cards := make([]model.Card, 0)
	for rows.Next() {
		var c model.Card
		if err := rows.Scan(&c.ID, &c.ListID, &c.Title, &c.Description, &c.Position, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan card: %w", err)
		}
		cards = append(cards, c)
	}

	return cards, rows.Err()
}

func (r *CardRepository) Update(ctx context.Context, id int64, title, description string) (*model.Card, error) {
	const query = `
		UPDATE cards
		SET title = $2, description = $3, updated_at = now()
		WHERE id = $1
		RETURNING id, list_id, title, description, position, created_at, updated_at
	`

	var c model.Card
	err := r.db.QueryRow(ctx, query, id, title, description).Scan(
		&c.ID, &c.ListID, &c.Title, &c.Description, &c.Position, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCardNotFound
		}
		return nil, fmt.Errorf("update card: %w", err)
	}

	return &c, nil
}

// UpdatePositionAndList sets a card's list and position. Intended to be
// called repeatedly inside a transaction when moving/resequencing cards.
func (r *CardRepository) UpdatePositionAndList(ctx context.Context, id, listID int64, position int) error {
	const query = `UPDATE cards SET list_id = $2, position = $3, updated_at = now() WHERE id = $1`

	if _, err := r.db.Exec(ctx, query, id, listID, position); err != nil {
		return fmt.Errorf("update card position: %w", err)
	}

	return nil
}

func (r *CardRepository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM cards WHERE id = $1`

	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete card: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrCardNotFound
	}

	return nil
}

package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"taskflow/internal/model"
)

var ErrListNotFound = errors.New("list not found")

type ListRepository struct {
	db Querier
}

func NewListRepository(db Querier) *ListRepository {
	return &ListRepository{db: db}
}

func (r *ListRepository) Create(ctx context.Context, boardID int64, name string, position int) (*model.List, error) {
	const query = `
		INSERT INTO lists (board_id, name, position)
		VALUES ($1, $2, $3)
		RETURNING id, board_id, name, position, created_at, updated_at
	`

	var l model.List
	err := r.db.QueryRow(ctx, query, boardID, name, position).Scan(
		&l.ID, &l.BoardID, &l.Name, &l.Position, &l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert list: %w", err)
	}

	return &l, nil
}

func (r *ListRepository) GetByID(ctx context.Context, id int64) (*model.List, error) {
	const query = `
		SELECT id, board_id, name, position, created_at, updated_at
		FROM lists
		WHERE id = $1
	`

	var l model.List
	err := r.db.QueryRow(ctx, query, id).Scan(
		&l.ID, &l.BoardID, &l.Name, &l.Position, &l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrListNotFound
		}
		return nil, fmt.Errorf("get list: %w", err)
	}

	return &l, nil
}

func (r *ListRepository) ListByBoard(ctx context.Context, boardID int64) ([]model.List, error) {
	const query = `
		SELECT id, board_id, name, position, created_at, updated_at
		FROM lists
		WHERE board_id = $1
		ORDER BY position ASC
	`

	rows, err := r.db.Query(ctx, query, boardID)
	if err != nil {
		return nil, fmt.Errorf("list lists: %w", err)
	}
	defer rows.Close()

	lists := make([]model.List, 0)
	for rows.Next() {
		var l model.List
		if err := rows.Scan(&l.ID, &l.BoardID, &l.Name, &l.Position, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan list: %w", err)
		}
		lists = append(lists, l)
	}

	return lists, rows.Err()
}

func (r *ListRepository) Update(ctx context.Context, id int64, name string) (*model.List, error) {
	const query = `
		UPDATE lists
		SET name = $2, updated_at = now()
		WHERE id = $1
		RETURNING id, board_id, name, position, created_at, updated_at
	`

	var l model.List
	err := r.db.QueryRow(ctx, query, id, name).Scan(
		&l.ID, &l.BoardID, &l.Name, &l.Position, &l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrListNotFound
		}
		return nil, fmt.Errorf("update list: %w", err)
	}

	return &l, nil
}

// UpdatePosition sets a single list's position. Intended to be called
// repeatedly inside a transaction when resequencing a board's lists.
func (r *ListRepository) UpdatePosition(ctx context.Context, id int64, position int) error {
	const query = `UPDATE lists SET position = $2, updated_at = now() WHERE id = $1`

	if _, err := r.db.Exec(ctx, query, id, position); err != nil {
		return fmt.Errorf("update list position: %w", err)
	}

	return nil
}

func (r *ListRepository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM lists WHERE id = $1`

	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete list: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrListNotFound
	}

	return nil
}

package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"taskflow/internal/model"
)

var ErrBoardNotFound = errors.New("board not found")

type BoardRepository struct {
	db Querier
}

func NewBoardRepository(db Querier) *BoardRepository {
	return &BoardRepository{db: db}
}

func (r *BoardRepository) Create(ctx context.Context, userID int64, name, description string) (*model.Board, error) {
	const query = `
		INSERT INTO boards (user_id, name, description)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, name, description, created_at, updated_at
	`

	var b model.Board
	err := r.db.QueryRow(ctx, query, userID, name, description).Scan(
		&b.ID, &b.UserID, &b.Name, &b.Description, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert board: %w", err)
	}

	return &b, nil
}

func (r *BoardRepository) GetByID(ctx context.Context, id int64) (*model.Board, error) {
	const query = `
		SELECT id, user_id, name, description, created_at, updated_at
		FROM boards
		WHERE id = $1
	`

	var b model.Board
	err := r.db.QueryRow(ctx, query, id).Scan(
		&b.ID, &b.UserID, &b.Name, &b.Description, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBoardNotFound
		}
		return nil, fmt.Errorf("get board: %w", err)
	}

	return &b, nil
}

func (r *BoardRepository) ListByUser(ctx context.Context, userID int64) ([]model.Board, error) {
	const query = `
		SELECT id, user_id, name, description, created_at, updated_at
		FROM boards
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list boards: %w", err)
	}
	defer rows.Close()

	boards := make([]model.Board, 0)
	for rows.Next() {
		var b model.Board
		if err := rows.Scan(&b.ID, &b.UserID, &b.Name, &b.Description, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan board: %w", err)
		}
		boards = append(boards, b)
	}

	return boards, rows.Err()
}

func (r *BoardRepository) Update(ctx context.Context, id int64, name, description string) (*model.Board, error) {
	const query = `
		UPDATE boards
		SET name = $2, description = $3, updated_at = now()
		WHERE id = $1
		RETURNING id, user_id, name, description, created_at, updated_at
	`

	var b model.Board
	err := r.db.QueryRow(ctx, query, id, name, description).Scan(
		&b.ID, &b.UserID, &b.Name, &b.Description, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBoardNotFound
		}
		return nil, fmt.Errorf("update board: %w", err)
	}

	return &b, nil
}

func (r *BoardRepository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM boards WHERE id = $1`

	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete board: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrBoardNotFound
	}

	return nil
}

package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

// model
type Todo struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"userID"`
	Title       string    `json:"title"`
	Desctiption string    `json:"description"`
	Completed   bool      `json:"completed"`
	Priority    int16     `json:"priority"`
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   string    `json:"updatedAt"`
}

type TodosStore struct {
	db *sql.DB
}

func (s *TodosStore) Create(ctx context.Context, todo *Todo) error {
	query := `
	 INSERT INTO todos (title, description, completed, priority, tags)
		VALUES ($1, $2, $3, $4, $5) RETURNING id, createdAt, updatedAt
	`
	err := s.db.QueryRowContext(
		ctx,
		query,
		todo.UserID,
		todo.Title,
		todo.Desctiption,
		todo.Completed,
		todo.Priority,
		pq.Array(todo.Tags),
	).Scan(
		&todo.ID,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)

	if err != nil {
		return err
	}
	return nil
}

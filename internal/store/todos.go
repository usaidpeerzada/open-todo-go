package store

import (
	"context"
	"database/sql"
)

type TodosStore struct {
	db *sql.DB
}

func (s *TodosStore) Create(ctx context.Context) error {
	return nil
}

package store

import (
	"context"
	"database/sql"
)

type Storage struct {
	Todos interface {
		Create(context.Context) error
	}
	Users interface {
		Create(context.Context) error
	}
}

func NewPostgresStorage(db *sql.DB) Storage {
	return Storage{
		Todos: &TodosStore{db},
		Users: &UserStore{db},
	}
}

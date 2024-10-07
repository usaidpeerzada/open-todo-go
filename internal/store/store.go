package store

import (
	"context"
	"database/sql"
)

type Storage struct {
	Todos interface {
		Create(context.Context, *Todo) error
	}
	Users interface {
		Create(context.Context, *User) error
	}
}

func NewPostgresStorage(db *sql.DB) Storage {
	return Storage{
		Todos: &TodosStore{db},
		Users: &UserStore{db},
	}
}

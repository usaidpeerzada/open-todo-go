package store

import (
	"context"
	"database/sql"
)

type Storage struct {
	Todos interface {
		Create(context.Context, *Todo) error
		GetAllTodos(context.Context, int64) ([]Todo, error)
		GetTodoByID(context.Context, int64) (*Todo, error)
		UpdateTodo(context.Context, int64, map[string]interface{}) error
		GetTodosByTag(context.Context, int64, string) ([]Todo, error)
		DeleteTodo(context.Context, int64) error
	}
	Users interface {
		Create(context.Context, *User) error
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Todos: &TodosStore{db},
		Users: &UserStore{db},
	}
}

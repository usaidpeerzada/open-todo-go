package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
)

// model
type Todo struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"userID"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	Priority    int16     `json:"priority"`
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   string    `json:"updatedAt"`
}

type TodosStore struct {
	db *sql.DB
}

// create todo
func (s *TodosStore) Create(ctx context.Context, todo *Todo) error {
	fmt.Println("what is this", todo)
	// pq: got 6 parameters but the statement requires 5, had to add user id here, but why?
	query := `
	 INSERT INTO todos (user_id, title, description, completed, priority, tags)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at, updated_at
	`
	err := s.db.QueryRowContext(
		ctx,
		query,
		todo.UserID,
		todo.Title,
		todo.Description,
		todo.Completed,
		todo.Priority,
		pq.Array(todo.Tags),
	).Scan(
		&todo.ID,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)
	fmt.Println(err)
	if err != nil {
		return err
	}
	return nil
}

// get all todos
func (s *TodosStore) GetAllTodos(ctx context.Context, userID int64) ([]Todo, error) {
	query := `
      SELECT id, user_id, title, description, completed, priority, tags, created_at, updated_at
      FROM todos
      WHERE user_id = $1
      ORDER BY created_at DESC
      `

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		err := rows.Scan(
			&todo.ID,
			&todo.UserID,
			&todo.Title,
			&todo.Description,
			&todo.Completed,
			&todo.Priority,
			pq.Array(&todo.Tags),
			&todo.CreatedAt,
			&todo.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return todos, nil
}

func (s *TodosStore) GetTodoByID(ctx context.Context, todoID int64) (*Todo, error) {
	query := `
    SELECT id, user_id, title, description, completed, priority, tags, created_at, updated_at
    FROM todos
    WHERE id = $1
    `
	var todo Todo
	err := s.db.QueryRowContext(ctx, query, todoID).Scan(
		&todo.ID, &todo.UserID, &todo.Title, &todo.Description, &todo.Completed,
		&todo.Priority, pq.Array(&todo.Tags), &todo.CreatedAt, &todo.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no todo found with ID %d", todoID)
	}
	if err != nil {
		return nil, err
	}
	return &todo, nil
}

func (s *TodosStore) UpdateTodo(ctx context.Context, todoID int64, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return fmt.Errorf("no fields to update")
	}

	// Prepare the query parts
	var queryFields []string
	var args []interface{}
	argCounter := 1

	for field, value := range updates {
		// Use double quotes for field names and $n placeholders for values
		queryFields = append(queryFields, fmt.Sprintf(`"%s" = $%d`, field, argCounter))
		fmt.Println(fmt.Sprintf(`"%s" = $%d`, field, argCounter))
		args = append(args, value)
		argCounter++
	}

	fmt.Println(strings.Join(queryFields, ", "), argCounter)
	// Construct the SQL query
	query := fmt.Sprintf("UPDATE todos SET %s WHERE id = $%d", strings.Join(queryFields, ", "), argCounter)

	// Append todoID as the final argument for the WHERE clause
	args = append(args, todoID)

	// Debugging: Print the query and arguments to verify correctness
	fmt.Println("Generated Query:", query)
	fmt.Println("Args:", args)

	// Execute the query
	_, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error updating todo: %w", err)
	}

	return nil
}

func (s *TodosStore) GetTodosByTag(ctx context.Context, userID int64, tag string) ([]Todo, error) {
	query := `
    SELECT id, user_id, title, description, completed, priority, tags, created_at, updated_at
    FROM todos
    WHERE user_id = $1 AND $2 = ANY(tags)
    ORDER BY created_at DESC
    `
	rows, err := s.db.QueryContext(ctx, query, userID, tag)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		err := rows.Scan(&todo.ID, &todo.UserID, &todo.Title, &todo.Description, &todo.Completed,
			&todo.Priority, pq.Array(&todo.Tags), &todo.CreatedAt, &todo.UpdatedAt)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}
	return todos, rows.Err()
}

func (s *TodosStore) DeleteTodo(ctx context.Context, todoID int64) error {
	query := `
        DELETE FROM todos
        WHERE id = $1
    `
	result, err := s.db.ExecContext(ctx, query, todoID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return err
	}

	return nil
}

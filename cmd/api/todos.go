package main

import (
	"context"
	"fmt"
	"net/http"
	"open-todo-go/internal/store"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/lib/pq"
)

type CreateTodoPayload struct {
	Title       string   `json:"title" validate:"required"`
	Description string   `json:"description"`
	Priority    int16    `json:"priority" validate:"min=0,max=5"`
	Completed   bool     `json:"completed"`
	Tags        []string `json:"tags"`
}

type UpdatedTodoPayload struct {
	Title       *string  `json:"title,omitempty"`
	Description *string  `json:"description,omitempty"`
	Priority    *int64   `json:"priority,omitempty"`
	Completed   *bool    `json:"completed,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

func (app *application) CreateTodo(w http.ResponseWriter, r *http.Request) {
	var payload CreateTodoPayload

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
	}

	userID := getUserIDFromContext(r.Context())

	todo := &store.Todo{
		UserID:      userID,
		Title:       payload.Title,
		Description: payload.Description,
		Priority:    payload.Priority,
		Completed:   payload.Completed,
		Tags:        payload.Tags,
	}
	if err := app.store.Todos.Create(r.Context(), todo); err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("failed to create todo: %w", err))
		return
	}

	return
}

func (app *application) GetAllTodos(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	todos, err := app.store.Todos.GetAllTodos(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to fetch todos", http.StatusInternalServerError)
		return
	}
	respondJSON(w, todos)
}

func (app *application) GetTodoById(w http.ResponseWriter, r *http.Request) {
	// userID := getUserIDFromContext(r.Context())
	idParam := chi.URLParam(r, "todoID")
	todoID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("failed to parse id: %w", err))
		return
	}

	todo, err := app.store.Todos.GetTodoByID(r.Context(), todoID)

	if err != nil {
		http.Error(w, "Failed to fetch todo id", http.StatusInternalServerError)
		return
	}
	respondJSON(w, todo)
}

func (app *application) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	todoID, err := strconv.ParseInt(chi.URLParam(r, "todoID"), 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("invalid todo ID: %w", err))
		return
	}

	var payload UpdatedTodoPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	updates := buildUpdatesMap(payload)

	if err := app.store.Todos.UpdateTodo(r.Context(), todoID, updates); err != nil {
		app.internalServerError(w, r, fmt.Errorf("failed to update todo: %w", err))
		return
	}

	app.jsonResponse(w, http.StatusOK, nil)
}

func (app *application) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	idParam, err := strconv.ParseInt(chi.URLParam(r, "todoID"), 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("invalid todo ID: %w", err))
		return
	}

	error := app.store.Todos.DeleteTodo(r.Context(), idParam)

	if error != nil {
		app.internalServerError(w, r, fmt.Errorf("failed to delete todo: %w", err))
		return
	}
	app.jsonResponse(w, http.StatusOK, nil)
}

// Helper function to build the updates map from the payload
func buildUpdatesMap(payload UpdatedTodoPayload) map[string]interface{} {
	updates := make(map[string]interface{})

	// Add fields to updates map if they are not nil
	if payload.Title != nil {
		updates["title"] = *payload.Title
	}
	if payload.Description != nil {
		updates["description"] = *payload.Description
	}
	if payload.Priority != nil {
		updates["priority"] = *payload.Priority
	}
	if payload.Completed != nil {
		updates["completed"] = *payload.Completed
	}
	if payload.Tags != nil {
		updates["tags"] = pq.Array(payload.Tags)
	}

	return updates
}

func getUserIDFromContext(ctx context.Context) int64 {
	// Implementation depends on your authentication system
	return 1 // Placeholder
}

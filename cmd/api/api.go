package main

import (
	"log"
	"net/http"
	"open-todo-go/internal/store"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type application struct {
	config config
	store  store.Storage
	logger *zap.SugaredLogger
}

type config struct {
	addr string
	db   dbConfig
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)
		r.Route("/todos", func(r chi.Router) {
			r.Get("/", app.GetAllTodos)
			r.Get("/{todoID}", app.GetTodoById)
			r.Post("/create", app.CreateTodo)
			r.Put("/update/{todoID}", app.UpdateTodo)
		})
		// r.Put("/todos/{id}", todoHandler.UpdateTodo)
		// r.Delete("/todos/{id}", todoHandler.DeleteTodo)
		// r.Get("/todos/tag/{tag}", todoHandler.GetTodosByTag)
	})
	return r
	// mux := http.NewServeMux()
	// mux.HandleFunc("GET /api/v1/get-todos", app.healthCheckHandler)
	// return mux
}

func (app *application) run(mux http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}
	log.Printf("Server is running on %s", app.config.addr)
	return srv.ListenAndServe()
}

package main

import (
	"expvar"
	"log"
	"net/http"
	"open-todo-go/internal/auth"
	"open-todo-go/internal/store"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type application struct {
	config        config
	store         store.Storage
	logger        *zap.SugaredLogger
	authenticator auth.Authenticator
}

type config struct {
	addr string
	db   dbConfig
	auth authConfig
}

type authConfig struct {
	basic basicConfig
	token tokenConfig
}

type tokenConfig struct {
	secret string
	exp    time.Duration
	iss    string
}

type basicConfig struct {
	user string
	pass string
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// r.Use(cors.Handler(cors.Options{
	// 	AllowedOrigins:   []string{env.GetString("CORS_ALLOWED_ORIGIN", "http://localhost:5174")},
	// 	AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	// 	AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
	// 	ExposedHeaders:   []string{"Link"},
	// 	AllowCredentials: false,
	// 	MaxAge:           300, // Maximum value not ignored by any of major browsers
	// }))

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)
		r.With(app.BasicAuthMiddleware()).Get("/debug/vars", expvar.Handler().ServeHTTP)
		r.Route("/todos", func(r chi.Router) {
			r.Use(app.AuthTokenMiddleware)
			r.Get("/", app.GetAllTodos)
			r.Get("/{todoID}", app.GetTodoById)
			// r.Get("/todos/tag/{tag}", todoHandler.GetTodosByTag)
			r.Post("/create", app.CreateTodo)
			r.Put("/update/{todoID}", app.UpdateTodo)
			r.Delete("/delete/{todoID}", app.DeleteTodo)
		})
		r.Route("/user", func(r chi.Router) {
			r.Post("/create", app.RegisterUserHandler)
			r.Post("/login", app.LoginHandler)
		})
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

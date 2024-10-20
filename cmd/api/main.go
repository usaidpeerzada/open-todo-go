package main

import (
	"log"
	"open-todo-go/internal/auth"
	"open-todo-go/internal/db"
	"open-todo-go/internal/env"
	"open-todo-go/internal/store"
	"time"

	"go.uber.org/zap"
)

func main() {
	cfg := config{
		addr: env.GetString("Addr", ":8080"),
		db: dbConfig{
			// addr:         env.GetString("DB_ADDR", ),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		auth: authConfig{
			basic: basicConfig{
				user: env.GetString("AUTH_BASIC_USER", "admin"),
				pass: env.GetString("AUTH_BASIC_PASS", "admin"),
			},
			token: tokenConfig{
				secret: env.GetString("AUTH_TOKEN_SECRET", ""),
				exp:    time.Hour * 24 * 3, // 3 days
				iss:    "open-todo-go",
			},
		},
	}
	db, err := db.New(cfg.db.addr, cfg.db.maxOpenConns, cfg.db.maxIdleConns, cfg.db.maxIdleTime)

	jwtAuthenticator := auth.NewJWTAuthenticator(
		cfg.auth.token.secret,
		cfg.auth.token.iss,
		cfg.auth.token.iss,
	)
	// Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	store := store.NewStorage(db)

	app := &application{
		config:        cfg,
		store:         store,
		authenticator: jwtAuthenticator,
		logger:        logger,
	}

	mux := app.mount()
	log.Fatal(app.run(mux))
}

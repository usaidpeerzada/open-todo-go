package main

import (
	"log"
	"open-todo-go/internal/env"
	"open-todo-go/internal/store"
)

func main() {
	cfg := config{
		addr: env.GetString("Addr", ":8080"),
	}

	store := store.NewPostgresStorage(nil)

	app := &application{
		config: cfg,
		store:  store,
	}

	mux := app.mount()
	log.Fatal(app.run(mux))
}

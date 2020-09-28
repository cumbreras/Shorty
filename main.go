package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cumbreras/shortener/ent"
	"github.com/cumbreras/shortener/repository"
	"github.com/cumbreras/shortener/server"
	"github.com/cumbreras/shortener/service"
	"github.com/gorilla/mux"
	"github.com/hashicorp/go-hclog"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
}

func run() error {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "Shortener",
		Level: hclog.LevelFromString("DEBUG"),
	})

	dbClient, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")

	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}

	defer dbClient.Close()

	if err := dbClient.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	rps := repository.New(dbClient, logger)
	svc := service.New(rps, logger)

	srv := server.New(mux.NewRouter(), logger, svc)

	s := &http.Server{
		Handler:      srv,
		Addr:         ":1337",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	return s.ListenAndServe()
}

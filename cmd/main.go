package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Netflix/go-env"

	"github.com/lualfe/offer-eligibility-api/internal/api"
	"github.com/lualfe/offer-eligibility-api/internal/config"
	"github.com/lualfe/offer-eligibility-api/internal/core"
	"github.com/lualfe/offer-eligibility-api/internal/repository/db"
)

func main() {
	var cfg config.Config
	_, err := env.UnmarshalFromEnviron(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	dbConn, err := cfg.Postgres.Conn()
	if err != nil {
		log.Fatal(err)
	}

	pingErr := dbConn.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}

	repository := db.NewRepo(dbConn)
	svc := core.NewService(repository)

	httpServer := api.NewServer(svc)
	handlers := httpServer.InitHandlers()

	server := &http.Server{
		Handler: handlers,
		Addr:    fmt.Sprintf(":%d", cfg.API.Port),
	}

	log.Print("starting server")
	log.Fatal(server.ListenAndServe())
}

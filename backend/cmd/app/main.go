package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"git.subcult.tv/subculture-collective/dsa-slate/internal/admin"
	appcfg "git.subcult.tv/subculture-collective/dsa-slate/internal/app"
	"git.subcult.tv/subculture-collective/dsa-slate/internal/directory"
	"git.subcult.tv/subculture-collective/dsa-slate/internal/httpapi"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required; set it or use docker compose to start the Postgres-backed API")
	}

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("create postgres pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("connect to postgres: %v", err)
	}

	var handler http.Handler
	adminToken := appcfg.AdminAPIToken()
	if adminToken != "" {
		handler = httpapi.NewServerWithAdmin(directory.NewPostgresStore(pool), admin.NewPostgresStore(pool), adminToken)
	} else {
		handler = httpapi.NewServer(directory.NewPostgresStore(pool))
	}
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

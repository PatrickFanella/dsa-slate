package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"git.subcult.tv/subculture-collective/dsa-slate/internal/admin"
	appcfg "git.subcult.tv/subculture-collective/dsa-slate/internal/app"
	"git.subcult.tv/subculture-collective/dsa-slate/internal/auth"
	"git.subcult.tv/subculture-collective/dsa-slate/internal/civic"
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

	databaseURL := appcfg.DatabaseURL()
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
	var civicService civic.LookupService
	if civicAPIKey := appcfg.GoogleCivicAPIKey(); civicAPIKey != "" {
		civicService = civic.NewGoogleClient(civicAPIKey, http.DefaultClient)
	}
	adminToken := appcfg.AdminAPIToken()
	authStore := auth.NewPostgresStore(pool)
	var googleAuth auth.GoogleSignInService
	if clientID, clientSecret := appcfg.GoogleOAuthClientID(), appcfg.GoogleOAuthClientSecret(); clientID != "" && clientSecret != "" {
		googleAuth = auth.NewGoogleSignInService(authStore, auth.NewGoogleOAuthConfig(clientID, clientSecret, appcfg.GoogleOAuthRedirectURL(), appcfg.AppBaseURL()), nil)
	}
	if adminToken != "" {
		handler = httpapi.NewServerWithAuthAndServices(directory.NewPostgresStore(pool), authStore, googleAuth, admin.NewPostgresStore(pool), adminToken, civicService)
	} else {
		handler = httpapi.NewServerWithAuthAndServices(directory.NewPostgresStore(pool), authStore, googleAuth, nil, "", civicService)
	}
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

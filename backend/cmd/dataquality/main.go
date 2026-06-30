package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	appcfg "git.subcult.tv/subculture-collective/dsa-slate/internal/app"
	"git.subcult.tv/subculture-collective/dsa-slate/internal/directory"
)

func main() {
	ctx := context.Background()
	databaseURL := appcfg.DatabaseURL()
	if databaseURL == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is required")
		os.Exit(2)
	}

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect database: %v\n", err)
		os.Exit(2)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "connect database: %v\n", err)
		os.Exit(2)
	}

	store := directory.NewPostgresStore(pool)
	data, err := store.Directory(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load directory: %v\n", err)
		os.Exit(2)
	}

	referenceDate := os.Getenv("QUALITY_REFERENCE_DATE")
	if referenceDate == "" {
		referenceDate = time.Now().UTC().Format("2006-01-02")
	}

	report := directory.EvaluateQuality(data, directory.QualityOptions{ReferenceDate: referenceDate, StaleAfterDays: 180})
	fmt.Print(report.String())
	if report.HasFailures() {
		os.Exit(1)
	}
}

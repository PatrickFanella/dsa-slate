package app

import (
	"net/url"
	"os"
)

const serviceName = "dsa-slate"

func ServiceName() string {
	return serviceName
}

func AdminAPIToken() string {
	return os.Getenv("ADMIN_API_TOKEN")
}

func DatabaseURL() string {
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		return databaseURL
	}

	user := envOrDefault("POSTGRES_USER", "dsa_slate")
	password := envOrDefault("POSTGRES_PASSWORD", "change-me")
	host := envOrDefault("DATABASE_HOST", "localhost")
	port := envOrDefault("DATABASE_PORT", envOrDefault("POSTGRES_PORT", "5432"))
	database := envOrDefault("POSTGRES_DB", "dsa_slate")

	postgresURL := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, password),
		Host:   host + ":" + port,
		Path:   "/" + database,
	}
	query := postgresURL.Query()
	query.Set("sslmode", "disable")
	postgresURL.RawQuery = query.Encode()

	return postgresURL.String()
}

func envOrDefault(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func GoogleCivicAPIKey() string {
	return os.Getenv("GOOGLE_CIVIC_API_KEY")
}

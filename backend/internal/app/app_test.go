package app

import "testing"

func TestServiceName(t *testing.T) {
	if got := ServiceName(); got != "dsa-slate" {
		t.Fatalf("unexpected service name: %s", got)
	}
}

func TestAppBaseURLDefault(t *testing.T) {
	t.Setenv("APP_BASE_URL", "")
	if got := AppBaseURL(); got != "http://localhost:5173" {
		t.Fatalf("AppBaseURL() = %q", got)
	}
}

func TestGoogleOAuthRedirectURLDefaultsToAppBaseURL(t *testing.T) {
	t.Setenv("APP_BASE_URL", "https://example.org/")
	t.Setenv("GOOGLE_OAUTH_REDIRECT_URL", "")

	if got := GoogleOAuthRedirectURL(); got != "https://example.org/api/auth/google/callback" {
		t.Fatalf("GoogleOAuthRedirectURL() = %q", got)
	}
}

func TestGoogleOAuthClientEnvHelpers(t *testing.T) {
	t.Setenv("GOOGLE_OAUTH_CLIENT_ID", "client-id")
	t.Setenv("GOOGLE_OAUTH_CLIENT_SECRET", "client-secret")

	if got := GoogleOAuthClientID(); got != "client-id" {
		t.Fatalf("GoogleOAuthClientID() = %q", got)
	}
	if got := GoogleOAuthClientSecret(); got != "client-secret" {
		t.Fatalf("GoogleOAuthClientSecret() = %q", got)
	}
}

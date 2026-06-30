package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestBuildGoogleAuthURLIncludesOIDCAndPKCE(t *testing.T) {
	cfg := NewGoogleOAuthConfig("client-id", "client-secret", "https://example.org/api/auth/google/callback", "https://example.org")
	state := "state-token"
	nonce := "nonce-token"
	verifier := "verifier-token"

	got := BuildGoogleAuthURL(cfg, state, nonce, verifier)
	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	query := parsed.Query()
	if query.Get("response_type") != "code" {
		t.Fatalf("response_type = %q, want code", query.Get("response_type"))
	}
	if query.Get("client_id") != "client-id" {
		t.Fatalf("client_id = %q, want client-id", query.Get("client_id"))
	}
	if query.Get("redirect_uri") != "https://example.org/api/auth/google/callback" {
		t.Fatalf("redirect_uri = %q, want callback URL", query.Get("redirect_uri"))
	}
	if query.Get("state") != state {
		t.Fatalf("state = %q, want %q", query.Get("state"), state)
	}
	if query.Get("nonce") != nonce {
		t.Fatalf("nonce = %q, want %q", query.Get("nonce"), nonce)
	}
	if query.Get("code_challenge_method") != "S256" {
		t.Fatalf("code_challenge_method = %q, want S256", query.Get("code_challenge_method"))
	}
	if query.Get("code_challenge") == "" {
		t.Fatal("code_challenge missing")
	}
	for _, scope := range []string{"openid", "email", "profile"} {
		if !strings.Contains(query.Get("scope"), scope) {
			t.Fatalf("scope %q missing from %q", scope, query.Get("scope"))
		}
	}
}

func TestTemporaryOAuthCookieHelpers(t *testing.T) {
	cookie := NewTemporaryOAuthCookie("cookie-name", "value", false)
	if !cookie.HttpOnly {
		t.Fatal("cookie should be HTTP-only")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("SameSite = %v, want Lax", cookie.SameSite)
	}
	if cookie.Path != "/" {
		t.Fatalf("Path = %q, want /", cookie.Path)
	}

	cleared := ClearTemporaryOAuthCookie("cookie-name", false)
	if cleared.MaxAge >= 0 || !cleared.Expires.Equal(time.Unix(0, 0).UTC()) {
		t.Fatalf("cleared cookie = %#v", cleared)
	}
}

func TestGoogleUserIDStable(t *testing.T) {
	if got := googleUserID("abc123"); got != "google:abc123" {
		t.Fatalf("googleUserID() = %q", got)
	}
}

func TestGoogleSignInCallbackSuccess(t *testing.T) {
	now := time.Date(2026, time.June, 29, 12, 0, 0, 0, time.UTC)
	store := &memoryAuthStore{
		users: map[string]User{
			"user-1": {ID: "user-1", Email: "editor@example.org", Role: RoleEditor},
		},
		identities: map[string]OAuthIdentity{
			googleOAuthProvider + "\x00" + "google-sub": {Provider: googleOAuthProvider, ProviderSubject: "google-sub", UserID: "user-1"},
		},
	}
	flow := &fakeOAuth2Flow{token: oauth2TokenWithIDToken("raw-id-token")}
	verifier := &fakeGoogleTokenVerifier{userInfo: GoogleUserInfo{Subject: "google-sub", Email: "editor@example.org", Name: "Ada Lovelace", EmailVerified: true, Nonce: "nonce-token"}}
	service := &googleSignInService{
		store:           store,
		cfg:             GoogleOAuthConfig{ClientID: "client-id", ClientSecret: "client-secret", RedirectURL: "https://example.org/api/auth/google/callback", AppBaseURL: "https://example.org", OAuth2: flow},
		verifier:        verifier,
		now:             func() time.Time { return now },
		sessionDuration: defaultGoogleSessionDuration,
		secureCookies:   true,
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/google/callback?state=state-token&code=code-token", nil)
	req.AddCookie(NewTemporaryOAuthCookie(googleOAuthStateCookieName, "state-token", true))
	req.AddCookie(NewTemporaryOAuthCookie(googleOAuthNonceCookieName, "nonce-token", true))
	req.AddCookie(NewTemporaryOAuthCookie(googleOAuthPKCECookieName, "verifier-token", true))

	service.Callback(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if got := rec.Header().Get("Location"); got != "/" {
		t.Fatalf("Location = %q, want /", got)
	}
	if !verifier.called || verifier.audience != "client-id" {
		t.Fatalf("verifier called=%v audience=%q", verifier.called, verifier.audience)
	}
	if flow.exchangeCode != "code-token" {
		t.Fatalf("exchangeCode = %q, want code-token", flow.exchangeCode)
	}
	if len(store.sessions) == 0 {
		t.Fatal("session was not created")
	}
	if user, ok := store.users["user-1"]; !ok || user.Role != RoleEditor || user.DisplayName != "Ada Lovelace" {
		t.Fatalf("user = %#v, want preserved editor role and updated display name", user)
	}
	if identity, ok := store.identities[googleOAuthProvider+"\x00"+"google-sub"]; !ok || identity.UserID != "user-1" {
		t.Fatalf("identity = %#v, want linked to user-1", identity)
	}
	setCookies := strings.Join(rec.Header().Values("Set-Cookie"), "\n")
	for _, name := range []string{googleOAuthStateCookieName, googleOAuthNonceCookieName, googleOAuthPKCECookieName, SessionCookieName} {
		if !strings.Contains(setCookies, name+"=") {
			t.Fatalf("missing cookie %s in %q", name, setCookies)
		}
	}
}

func TestGoogleSignInCallbackRejectsPrivilegedEmailAutoLink(t *testing.T) {
	store := &memoryAuthStore{
		users: map[string]User{
			"admin-1": {ID: "admin-1", Email: "admin@example.org", Role: RoleAdmin},
		},
	}
	service := &googleSignInService{
		store:           store,
		cfg:             GoogleOAuthConfig{ClientID: "client-id", ClientSecret: "client-secret", RedirectURL: "https://example.org/api/auth/google/callback", AppBaseURL: "https://example.org", OAuth2: &fakeOAuth2Flow{token: oauth2TokenWithIDToken("raw-id-token")}},
		verifier:        &fakeGoogleTokenVerifier{userInfo: GoogleUserInfo{Subject: "new-google-sub", Email: "admin@example.org", Name: "Admin", EmailVerified: true, Nonce: "nonce-token"}},
		now:             time.Now,
		sessionDuration: defaultGoogleSessionDuration,
		secureCookies:   true,
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/google/callback?state=state-token&code=code-token", nil)
	req.AddCookie(NewTemporaryOAuthCookie(googleOAuthStateCookieName, "state-token", true))
	req.AddCookie(NewTemporaryOAuthCookie(googleOAuthNonceCookieName, "nonce-token", true))
	req.AddCookie(NewTemporaryOAuthCookie(googleOAuthPKCECookieName, "verifier-token", true))

	service.Callback(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	if len(store.sessions) != 0 {
		t.Fatalf("sessions = %#v, want none", store.sessions)
	}
}

func TestGoogleSignInCallbackRejectsBadState(t *testing.T) {
	store := &memoryAuthStore{}
	flow := &fakeOAuth2Flow{token: oauth2TokenWithIDToken("raw-id-token")}
	service := &googleSignInService{
		store:         store,
		cfg:           GoogleOAuthConfig{ClientID: "client-id", ClientSecret: "client-secret", RedirectURL: "https://example.org/api/auth/google/callback", AppBaseURL: "https://example.org", OAuth2: flow},
		verifier:      &fakeGoogleTokenVerifier{userInfo: GoogleUserInfo{Subject: "google-sub", Email: "editor@example.org", EmailVerified: true, Nonce: "nonce-token"}},
		secureCookies: true,
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/google/callback?state=wrong&code=code-token", nil)
	req.AddCookie(NewTemporaryOAuthCookie(googleOAuthStateCookieName, "state-token", true))
	req.AddCookie(NewTemporaryOAuthCookie(googleOAuthNonceCookieName, "nonce-token", true))
	req.AddCookie(NewTemporaryOAuthCookie(googleOAuthPKCECookieName, "verifier-token", true))

	service.Callback(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if flow.exchangeCalled {
		t.Fatal("exchange should not be called for bad state")
	}
}

type fakeOAuth2Flow struct {
	exchangeCalled bool
	exchangeCode   string
	state          string
	token          *oauth2.Token
	url            string
	wrappedOpts    int
}

func (f *fakeOAuth2Flow) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	f.state = state
	f.wrappedOpts = len(opts)
	if f.url != "" {
		return f.url
	}
	return "https://accounts.example/auth?state=" + url.QueryEscape(state)
}

func (f *fakeOAuth2Flow) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	f.exchangeCalled = true
	f.exchangeCode = code
	f.wrappedOpts = len(opts)
	if f.token != nil {
		return f.token, nil
	}
	return oauth2TokenWithIDToken("raw-id-token"), nil
}

type fakeGoogleTokenVerifier struct {
	called    bool
	audience  string
	userInfo  GoogleUserInfo
	returnErr error
}

func (f *fakeGoogleTokenVerifier) Verify(ctx context.Context, rawIDToken string, audience string) (GoogleUserInfo, error) {
	f.called = true
	f.audience = audience
	if f.returnErr != nil {
		return GoogleUserInfo{}, f.returnErr
	}
	return f.userInfo, nil
}

func oauth2TokenWithIDToken(rawIDToken string) *oauth2.Token {
	return (&oauth2.Token{}).WithExtra(map[string]interface{}{"id_token": rawIDToken})
}

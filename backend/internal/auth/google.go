package auth

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
)

const (
	googleOAuthProvider          = "google"
	googleOAuthStateCookieName   = "dsa_google_oauth_state"
	googleOAuthNonceCookieName   = "dsa_google_oauth_nonce"
	googleOAuthPKCECookieName    = "dsa_google_oauth_pkce"
	googleOAuthCookiePath        = "/"
	googleOAuthCookieDuration    = 10 * time.Minute
	defaultGoogleSessionDuration = 7 * 24 * time.Hour
)

// OAuth2Flow is the subset of oauth2.Config used by the Google sign-in flow.
type OAuth2Flow interface {
	AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
}

// GoogleOAuthConfig holds the browser-facing and OAuth2 configuration.
type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	AppBaseURL   string
	Scopes       []string
	OAuth2       OAuth2Flow
}

// GoogleUserInfo is the normalized identity data extracted from a verified ID token.
type GoogleUserInfo struct {
	Subject       string
	Email         string
	Name          string
	EmailVerified bool
	Nonce         string
}

// GoogleTokenVerifier verifies a Google ID token.
type GoogleTokenVerifier interface {
	Verify(ctx context.Context, rawIDToken string, audience string) (GoogleUserInfo, error)
}

var ErrOAuthIdentityLinkRequired = errors.New("existing privileged user requires explicit oauth identity link")

// GoogleSignInService starts and completes the Google OIDC flow.
type GoogleSignInService interface {
	Start(http.ResponseWriter, *http.Request)
	Callback(http.ResponseWriter, *http.Request)
}

type googleTokenVerifier struct{}

// NewGoogleOAuthConfig creates a Google OAuth config with the required OIDC scopes.
func NewGoogleOAuthConfig(clientID, clientSecret, redirectURL, appBaseURL string) GoogleOAuthConfig {
	return GoogleOAuthConfig{
		ClientID:     strings.TrimSpace(clientID),
		ClientSecret: strings.TrimSpace(clientSecret),
		RedirectURL:  strings.TrimSpace(redirectURL),
		AppBaseURL:   strings.TrimSpace(appBaseURL),
		Scopes:       []string{"openid", "email", "profile"},
		OAuth2: &oauth2.Config{
			ClientID:     strings.TrimSpace(clientID),
			ClientSecret: strings.TrimSpace(clientSecret),
			RedirectURL:  strings.TrimSpace(redirectURL),
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
	}
}

// BuildGoogleAuthURL creates the Google authorization URL for the OIDC code flow.
func BuildGoogleAuthURL(cfg GoogleOAuthConfig, state, nonce, pkceVerifier string) string {
	if cfg.OAuth2 == nil {
		return ""
	}
	challenge := pkceChallenge(pkceVerifier)
	return cfg.OAuth2.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("nonce", nonce),
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)
}

// NewTemporaryOAuthCookie creates a short-lived HTTP-only cookie for the OAuth handshake.
func NewTemporaryOAuthCookie(name string, value string, secure bool) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     googleOAuthCookiePath,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(googleOAuthCookieDuration.Seconds()),
		Expires:  time.Now().UTC().Add(googleOAuthCookieDuration),
	}
}

// ClearTemporaryOAuthCookie returns a deletion cookie for the OAuth handshake state.
func ClearTemporaryOAuthCookie(name string, secure bool) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     googleOAuthCookiePath,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0).UTC(),
	}
}

// VerifyGoogleIDToken verifies a Google ID token and extracts the normalized user info.
func VerifyGoogleIDToken(ctx context.Context, rawIDToken string, audience string) (GoogleUserInfo, error) {
	payload, err := idtoken.Validate(ctx, rawIDToken, audience)
	if err != nil {
		return GoogleUserInfo{}, fmt.Errorf("validate google id token: %w", err)
	}
	return googleUserInfoFromPayload(payload)
}

func (googleTokenVerifier) Verify(ctx context.Context, rawIDToken string, audience string) (GoogleUserInfo, error) {
	return VerifyGoogleIDToken(ctx, rawIDToken, audience)
}

// NewGoogleSignInService builds the Google OIDC sign-in flow handler.
func NewGoogleSignInService(store Store, cfg GoogleOAuthConfig, verifier GoogleTokenVerifier) GoogleSignInService {
	if verifier == nil {
		verifier = googleTokenVerifier{}
	}
	if cfg.OAuth2 == nil {
		cfg = NewGoogleOAuthConfig(cfg.ClientID, cfg.ClientSecret, cfg.RedirectURL, cfg.AppBaseURL)
	}
	return &googleSignInService{
		store:           store,
		cfg:             cfg,
		verifier:        verifier,
		now:             time.Now,
		sessionDuration: defaultGoogleSessionDuration,
		secureCookies:   strings.HasPrefix(strings.TrimSpace(cfg.AppBaseURL), "https://"),
	}
}

type googleSignInService struct {
	store           Store
	cfg             GoogleOAuthConfig
	verifier        GoogleTokenVerifier
	now             func() time.Time
	sessionDuration time.Duration
	secureCookies   bool
}

// Start begins the Google authorization code flow.
func (s *googleSignInService) Start(w http.ResponseWriter, r *http.Request) {
	if !s.configured() {
		writeAuthJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "auth_not_configured"})
		return
	}

	state, err := GenerateSessionToken()
	if err != nil {
		writeAuthJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_server_error"})
		return
	}
	nonce, err := GenerateSessionToken()
	if err != nil {
		writeAuthJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_server_error"})
		return
	}
	verifier, err := GenerateSessionToken()
	if err != nil {
		writeAuthJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_server_error"})
		return
	}

	http.SetCookie(w, NewTemporaryOAuthCookie(googleOAuthStateCookieName, state, s.secureCookies))
	http.SetCookie(w, NewTemporaryOAuthCookie(googleOAuthNonceCookieName, nonce, s.secureCookies))
	http.SetCookie(w, NewTemporaryOAuthCookie(googleOAuthPKCECookieName, verifier, s.secureCookies))

	http.Redirect(w, r, BuildGoogleAuthURL(s.cfg, state, nonce, verifier), http.StatusFound)
}

// Callback completes the Google authorization code flow.
func (s *googleSignInService) Callback(w http.ResponseWriter, r *http.Request) {
	if !s.configured() {
		writeAuthJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "auth_not_configured"})
		return
	}
	fail := func(status int, payload any) {
		s.clearOAuthCookies(w)
		writeAuthJSON(w, status, payload)
	}

	state := strings.TrimSpace(r.URL.Query().Get("state"))
	code := strings.TrimSpace(r.URL.Query().Get("code"))
	if state == "" || code == "" {
		fail(http.StatusBadRequest, map[string]any{"error": "validation_failed", "messages": []string{"state and code are required"}})
		return
	}

	stateCookie, err := r.Cookie(googleOAuthStateCookieName)
	if err != nil || stateCookie.Value == "" || stateCookie.Value != state {
		fail(http.StatusBadRequest, map[string]any{"error": "validation_failed", "messages": []string{"state mismatch"}})
		return
	}
	verifierCookie, err := r.Cookie(googleOAuthPKCECookieName)
	if err != nil || strings.TrimSpace(verifierCookie.Value) == "" {
		fail(http.StatusBadRequest, map[string]any{"error": "validation_failed", "messages": []string{"missing PKCE verifier"}})
		return
	}
	nonceCookie, err := r.Cookie(googleOAuthNonceCookieName)
	if err != nil || strings.TrimSpace(nonceCookie.Value) == "" {
		fail(http.StatusBadRequest, map[string]any{"error": "validation_failed", "messages": []string{"missing nonce"}})
		return
	}

	token, err := s.cfg.OAuth2.Exchange(r.Context(), code, oauth2.SetAuthURLParam("code_verifier", verifierCookie.Value))
	if err != nil {
		fail(http.StatusBadGateway, map[string]string{"error": "auth_exchange_failed"})
		return
	}
	rawIDToken, _ := token.Extra("id_token").(string)
	if strings.TrimSpace(rawIDToken) == "" {
		fail(http.StatusBadGateway, map[string]string{"error": "auth_exchange_failed"})
		return
	}

	info, err := s.verifier.Verify(r.Context(), rawIDToken, s.cfg.ClientID)
	if err != nil {
		fail(http.StatusBadGateway, map[string]string{"error": "auth_verification_failed"})
		return
	}
	if info.Nonce == "" || info.Nonce != nonceCookie.Value {
		fail(http.StatusBadRequest, map[string]any{"error": "validation_failed", "messages": []string{"nonce mismatch"}})
		return
	}
	if info.Subject == "" || info.Email == "" || !info.EmailVerified {
		fail(http.StatusBadRequest, map[string]any{"error": "validation_failed", "messages": []string{"verified email is required"}})
		return
	}

	user, err := s.resolveUser(r.Context(), info)
	if err != nil {
		if errors.Is(err, ErrOAuthIdentityLinkRequired) {
			fail(http.StatusForbidden, map[string]string{"error": "auth_link_required"})
			return
		}
		fail(http.StatusInternalServerError, map[string]string{"error": "internal_server_error"})
		return
	}

	_, err = s.store.UpsertOAuthIdentity(r.Context(), OAuthIdentity{
		Provider:        googleOAuthProvider,
		ProviderSubject: info.Subject,
		UserID:          user.ID,
	})
	if err != nil {
		fail(http.StatusInternalServerError, map[string]string{"error": "internal_server_error"})
		return
	}

	sessionToken, err := GenerateSessionToken()
	if err != nil {
		fail(http.StatusInternalServerError, map[string]string{"error": "internal_server_error"})
		return
	}
	session, err := s.store.CreateSession(r.Context(), Session{
		ID:        sessionToken,
		UserID:    user.ID,
		ExpiresAt: s.now().UTC().Add(s.sessionDuration),
	})
	if err != nil {
		fail(http.StatusInternalServerError, map[string]string{"error": "internal_server_error"})
		return
	}

	s.clearOAuthCookies(w)
	http.SetCookie(w, NewSessionCookie(session.ID, session.ExpiresAt, s.secureCookies))
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *googleSignInService) configured() bool {
	return s != nil && s.store != nil && s.verifier != nil && s.cfg.OAuth2 != nil && s.cfg.ClientID != "" && s.cfg.ClientSecret != "" && s.cfg.RedirectURL != ""
}

func (s *googleSignInService) clearOAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, ClearTemporaryOAuthCookie(googleOAuthStateCookieName, s.secureCookies))
	http.SetCookie(w, ClearTemporaryOAuthCookie(googleOAuthNonceCookieName, s.secureCookies))
	http.SetCookie(w, ClearTemporaryOAuthCookie(googleOAuthPKCECookieName, s.secureCookies))
}

func (s *googleSignInService) resolveUser(ctx context.Context, info GoogleUserInfo) (User, error) {
	if identity, err := s.store.GetOAuthIdentity(ctx, googleOAuthProvider, info.Subject); err == nil {
		user, err := s.store.GetUser(ctx, identity.UserID)
		if err != nil {
			return User{}, err
		}
		user.Email = info.Email
		if strings.TrimSpace(info.Name) != "" {
			user.DisplayName = info.Name
		}
		return s.store.UpsertUser(ctx, user)
	} else if !errors.Is(err, sql.ErrNoRows) {
		return User{}, err
	}

	if user, err := s.store.GetUserByEmail(ctx, info.Email); err == nil {
		if user.Role != RoleUser {
			return User{}, ErrOAuthIdentityLinkRequired
		}
		user.Email = info.Email
		if strings.TrimSpace(info.Name) != "" {
			user.DisplayName = info.Name
		}
		return s.store.UpsertUser(ctx, user)
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return User{}, err
	}

	user := User{
		ID:          googleUserID(info.Subject),
		Email:       info.Email,
		DisplayName: info.Name,
		Role:        RoleUser,
	}
	return s.store.UpsertUser(ctx, user)
}

func googleUserID(subject string) string {
	return "google:" + strings.TrimSpace(subject)
}

func googleUserInfoFromPayload(payload *idtoken.Payload) (GoogleUserInfo, error) {
	if payload == nil {
		return GoogleUserInfo{}, fmt.Errorf("validate google id token: empty payload")
	}
	claims := payload.Claims
	info := GoogleUserInfo{
		Subject:       strings.TrimSpace(payload.Subject),
		Email:         claimString(claims, "email"),
		Name:          claimString(claims, "name"),
		EmailVerified: claimBool(claims, "email_verified"),
		Nonce:         claimString(claims, "nonce"),
	}
	if info.Subject == "" {
		return GoogleUserInfo{}, fmt.Errorf("validate google id token: missing subject")
	}
	return info, nil
}

func claimString(claims map[string]interface{}, key string) string {
	value, _ := claims[key]
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case json.Number:
		return typed.String()
	default:
		return ""
	}
}

func claimBool(claims map[string]interface{}, key string) bool {
	value, ok := claims[key]
	if !ok {
		return false
	}
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		return strings.EqualFold(strings.TrimSpace(typed), "true")
	default:
		return false
	}
}

func pkceChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

func writeAuthJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

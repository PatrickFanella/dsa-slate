package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// User describes a persisted user record.
type User struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName,omitempty"`
	Role        Role      `json:"role"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Session is an opaque server-side session record.
type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
}

// Summary is the minimal user payload returned to the browser.
type Summary struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  Role   `json:"role"`
}

// State describes the current auth session status.
type State struct {
	Authenticated bool     `json:"authenticated"`
	User          *Summary `json:"user,omitempty"`
}

// Store is the DB-facing contract used by auth/session wiring.
type Store interface {
	UpsertUser(context.Context, User) (User, error)
	GetUser(context.Context, string) (User, error)
	GetUserByEmail(context.Context, string) (User, error)
	CreateSession(context.Context, Session) (Session, error)
	GetSession(context.Context, string) (Session, error)
	DeleteSession(context.Context, string) error
}

// Active reports whether the session is still valid at the provided time.
func (s Session) Active(now time.Time) bool {
	return !s.Expired(now)
}

// Expired reports whether the session should no longer be accepted.
func (s Session) Expired(now time.Time) bool {
	return !now.Before(s.ExpiresAt)
}

// ResolveState loads the current session from the store and returns anonymous state on cache misses.
func ResolveState(ctx context.Context, store Store, token string, now time.Time) (State, error) {
	token = strings.TrimSpace(token)
	if token == "" || store == nil {
		return State{}, nil
	}

	session, err := store.GetSession(ctx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return State{}, nil
		}
		return State{}, fmt.Errorf("lookup session: %w", err)
	}
	if session.Expired(now) {
		return State{}, nil
	}

	user, err := store.GetUser(ctx, session.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return State{}, nil
		}
		return State{}, fmt.Errorf("lookup user: %w", err)
	}

	role, err := ParseRole(string(user.Role))
	if err != nil {
		return State{}, fmt.Errorf("validate role: %w", err)
	}

	return State{
		Authenticated: true,
		User: &Summary{
			ID:    user.ID,
			Email: user.Email,
			Role:  role,
		},
	}, nil
}

// NewSessionCookie builds the future session cookie shape.
func NewSessionCookie(token string, expiresAt time.Time, secure bool) *http.Cookie {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}
	return &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
		MaxAge:   maxAge,
	}
}

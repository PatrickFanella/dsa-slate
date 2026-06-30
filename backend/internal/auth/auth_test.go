package auth

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"
)

func TestValidateRole(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
		want    Role
	}{
		{name: "default user", value: "", want: RoleUser},
		{name: "user", value: "user", want: RoleUser},
		{name: "editor", value: "editor", want: RoleEditor},
		{name: "admin", value: "admin", want: RoleAdmin},
		{name: "invalid", value: "superuser", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRole(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Fatal("ParseRole() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseRole() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("ParseRole() = %q, want %q", got, tt.want)
			}
			if err := ValidateRole(tt.value); err != nil {
				t.Fatalf("ValidateRole() = %v, want nil", err)
			}
		})
	}
}

func TestSessionExpiryBehavior(t *testing.T) {
	now := time.Date(2026, time.June, 29, 12, 0, 0, 0, time.UTC)
	session := Session{ExpiresAt: now.Add(time.Minute)}
	if session.Expired(now) {
		t.Fatal("session expired too early")
	}
	if !session.Active(now) {
		t.Fatal("session should be active")
	}

	expired := Session{ExpiresAt: now}
	if !expired.Expired(now) {
		t.Fatal("session expiring at now should be expired")
	}
}

func TestGenerateSessionTokenShapeAndUniqueness(t *testing.T) {
	pattern := regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)
	seen := make(map[string]struct{}, 256)
	for i := 0; i < 256; i++ {
		token, err := GenerateSessionToken()
		if err != nil {
			t.Fatalf("GenerateSessionToken() error = %v", err)
		}
		if !pattern.MatchString(token) {
			t.Fatalf("token %q has unexpected shape", token)
		}
		if _, ok := seen[token]; ok {
			t.Fatalf("duplicate token generated: %q", token)
		}
		seen[token] = struct{}{}
	}
}

func TestResolveState(t *testing.T) {
	now := time.Date(2026, time.June, 29, 12, 0, 0, 0, time.UTC)
	store := &memoryAuthStore{
		sessions: map[string]Session{
			"valid-token":   {ID: "valid-token", UserID: "user-1", ExpiresAt: now.Add(time.Hour)},
			"expired-token": {ID: "expired-token", UserID: "user-1", ExpiresAt: now.Add(-time.Minute)},
		},
		users: map[string]User{
			"user-1": {ID: "user-1", Email: "editor@example.org", Role: RoleEditor},
		},
	}

	tests := []struct {
		name     string
		token    string
		wantAuth bool
	}{
		{name: "no token", token: ""},
		{name: "missing session", token: "missing"},
		{name: "expired", token: "expired-token"},
		{name: "valid", token: "valid-token", wantAuth: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, err := ResolveState(context.Background(), store, tt.token, now)
			if err != nil {
				t.Fatalf("ResolveState() error = %v", err)
			}
			if state.Authenticated != tt.wantAuth {
				t.Fatalf("Authenticated = %v, want %v", state.Authenticated, tt.wantAuth)
			}
			if tt.wantAuth {
				if state.User == nil || state.User.ID != "user-1" || state.User.Email != "editor@example.org" || state.User.Role != RoleEditor {
					t.Fatalf("user = %#v, want user-1/editor@example.org/editor", state.User)
				}
			}
		})
	}
}

type memoryAuthStore struct {
	users    map[string]User
	sessions map[string]Session
}

func (s *memoryAuthStore) UpsertUser(ctx context.Context, user User) (User, error) {
	if s.users == nil {
		s.users = make(map[string]User)
	}
	s.users[user.ID] = user
	return user, nil
}

func (s *memoryAuthStore) GetUser(ctx context.Context, id string) (User, error) {
	if user, ok := s.users[id]; ok {
		return user, nil
	}
	return User{}, sql.ErrNoRows
}

func (s *memoryAuthStore) GetUserByEmail(ctx context.Context, email string) (User, error) {
	for _, user := range s.users {
		if user.Email == email {
			return user, nil
		}
	}
	return User{}, sql.ErrNoRows
}

func (s *memoryAuthStore) CreateSession(ctx context.Context, session Session) (Session, error) {
	if s.sessions == nil {
		s.sessions = make(map[string]Session)
	}
	s.sessions[session.ID] = session
	return session, nil
}

func (s *memoryAuthStore) GetSession(ctx context.Context, token string) (Session, error) {
	if session, ok := s.sessions[token]; ok {
		return session, nil
	}
	return Session{}, sql.ErrNoRows
}

func (s *memoryAuthStore) DeleteSession(ctx context.Context, token string) error {
	delete(s.sessions, token)
	return nil
}

package auth

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresStore persists auth users and sessions in PostgreSQL.
type PostgresStore struct {
	pool *pgxpool.Pool
}

// NewPostgresStore creates a PostgreSQL-backed auth store.
func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

// UpsertUser inserts or updates a user record.
func (s *PostgresStore) UpsertUser(ctx context.Context, user User) (User, error) {
	role, err := ParseRole(string(user.Role))
	if err != nil {
		return User{}, err
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO users (id, email, display_name, role)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			email = EXCLUDED.email,
			display_name = EXCLUDED.display_name,
			role = EXCLUDED.role,
			updated_at = now()
		RETURNING id, email, display_name, role, created_at, updated_at
	`, strings.TrimSpace(user.ID), strings.TrimSpace(user.Email), strings.TrimSpace(user.DisplayName), role)
	return scanUser(row)
}

// GetUser fetches a user by ID.
func (s *PostgresStore) GetUser(ctx context.Context, id string) (User, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, email, display_name, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`, strings.TrimSpace(id))
	return scanUser(row)
}

// GetUserByEmail fetches a user by email address.
func (s *PostgresStore) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, email, display_name, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`, strings.TrimSpace(email))
	return scanUser(row)
}

// UpsertOAuthIdentity inserts or updates an external OAuth identity record.
func (s *PostgresStore) UpsertOAuthIdentity(ctx context.Context, identity OAuthIdentity) (OAuthIdentity, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO user_oauth_identities (provider, provider_subject, user_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (provider, provider_subject) DO UPDATE SET
			user_id = EXCLUDED.user_id
		RETURNING provider, provider_subject, user_id, created_at
	`, strings.TrimSpace(identity.Provider), strings.TrimSpace(identity.ProviderSubject), strings.TrimSpace(identity.UserID))
	return scanOAuthIdentity(row)
}

// GetOAuthIdentity fetches an OAuth identity by provider and subject.
func (s *PostgresStore) GetOAuthIdentity(ctx context.Context, provider string, subject string) (OAuthIdentity, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT provider, provider_subject, user_id, created_at
		FROM user_oauth_identities
		WHERE provider = $1 AND provider_subject = $2
	`, strings.TrimSpace(provider), strings.TrimSpace(subject))
	return scanOAuthIdentity(row)
}

// CreateSession inserts or updates an opaque session record.
func (s *PostgresStore) CreateSession(ctx context.Context, session Session) (Session, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO user_sessions (id, user_id, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			expires_at = EXCLUDED.expires_at
		RETURNING id, user_id, expires_at, created_at
	`, strings.TrimSpace(session.ID), strings.TrimSpace(session.UserID), session.ExpiresAt)
	return scanSession(row)
}

// GetSession fetches a session by opaque ID.
func (s *PostgresStore) GetSession(ctx context.Context, token string) (Session, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, user_id, expires_at, created_at
		FROM user_sessions
		WHERE id = $1
	`, strings.TrimSpace(token))
	return scanSession(row)
}

// DeleteSession deletes an opaque session by ID.
func (s *PostgresStore) DeleteSession(ctx context.Context, token string) error {
	if _, err := s.pool.Exec(ctx, `DELETE FROM user_sessions WHERE id = $1`, strings.TrimSpace(token)); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

type rowScanner interface {
	Scan(...any) error
}

func scanUser(row rowScanner) (User, error) {
	var (
		id          sql.NullString
		email       sql.NullString
		displayName sql.NullString
		roleValue   sql.NullString
		createdAt   sql.NullTime
		updatedAt   sql.NullTime
	)
	if err := row.Scan(&id, &email, &displayName, &roleValue, &createdAt, &updatedAt); err != nil {
		return User{}, fmt.Errorf("scan user: %w", err)
	}
	role, err := ParseRole(roleValue.String)
	if err != nil {
		return User{}, fmt.Errorf("scan user role: %w", err)
	}
	return User{
		ID:          id.String,
		Email:       email.String,
		DisplayName: displayName.String,
		Role:        role,
		CreatedAt:   createdAt.Time,
		UpdatedAt:   updatedAt.Time,
	}, nil
}

func scanSession(row rowScanner) (Session, error) {
	var (
		id        sql.NullString
		userID    sql.NullString
		expiresAt sql.NullTime
		createdAt sql.NullTime
	)
	if err := row.Scan(&id, &userID, &expiresAt, &createdAt); err != nil {
		return Session{}, fmt.Errorf("scan session: %w", err)
	}
	return Session{
		ID:        id.String,
		UserID:    userID.String,
		ExpiresAt: expiresAt.Time,
		CreatedAt: createdAt.Time,
	}, nil
}

func scanOAuthIdentity(row rowScanner) (OAuthIdentity, error) {
	var (
		provider        sql.NullString
		providerSubject sql.NullString
		userID          sql.NullString
		createdAt       sql.NullTime
	)
	if err := row.Scan(&provider, &providerSubject, &userID, &createdAt); err != nil {
		return OAuthIdentity{}, fmt.Errorf("scan oauth identity: %w", err)
	}
	return OAuthIdentity{
		Provider:        provider.String,
		ProviderSubject: providerSubject.String,
		UserID:          userID.String,
		CreatedAt:       createdAt.Time,
	}, nil
}

var _ Store = (*PostgresStore)(nil)

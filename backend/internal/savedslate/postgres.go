package savedslate

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresStore persists saved slate items in PostgreSQL.
type PostgresStore struct {
	pool *pgxpool.Pool
}

// NewPostgresStore creates a PostgreSQL-backed saved slate store.
func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

// ListForUser returns saved slate items for one user, oldest first.
func (s *PostgresStore) ListForUser(ctx context.Context, userID string) ([]Item, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, endorsement_id, created_at, updated_at
		FROM saved_slate_items
		WHERE user_id = $1
		ORDER BY created_at ASC, id ASC
	`, strings.TrimSpace(userID))
	if err != nil {
		return nil, fmt.Errorf("list saved slate items: %w", err)
	}
	defer rows.Close()

	items := make([]Item, 0)
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list saved slate items: %w", err)
	}
	return items, nil
}

// UpsertForUser saves one public endorsement to a user's slate.
func (s *PostgresStore) UpsertForUser(ctx context.Context, userID string, endorsementID string) (Item, error) {
	userID = strings.TrimSpace(userID)
	endorsementID = strings.TrimSpace(endorsementID)
	if userID == "" || endorsementID == "" {
		return Item{}, ErrInvalidReference
	}

	var exists bool
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM endorsements WHERE id = $1)`, endorsementID).Scan(&exists); err != nil {
		return Item{}, fmt.Errorf("validate endorsement reference: %w", err)
	}
	if !exists {
		return Item{}, ErrInvalidReference
	}

	row := s.pool.QueryRow(ctx, `
		INSERT INTO saved_slate_items (id, user_id, endorsement_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, endorsement_id) DO UPDATE SET updated_at = now()
		RETURNING id, user_id, endorsement_id, created_at, updated_at
	`, itemID(userID, endorsementID), userID, endorsementID)
	return scanItem(row)
}

// DeleteForUser removes one saved slate item owned by the given user.
func (s *PostgresStore) DeleteForUser(ctx context.Context, userID string, itemID string) error {
	result, err := s.pool.Exec(ctx, `
		DELETE FROM saved_slate_items
		WHERE user_id = $1 AND id = $2
	`, strings.TrimSpace(userID), strings.TrimSpace(itemID))
	if err != nil {
		return fmt.Errorf("delete saved slate item: %w", err)
	}
	if result.RowsAffected() == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func itemID(userID string, endorsementID string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(userID) + "\x00" + strings.TrimSpace(endorsementID)))
	return "ssi_" + hex.EncodeToString(sum[:])
}

type rowScanner interface {
	Scan(...any) error
}

func scanItem(row rowScanner) (Item, error) {
	var (
		id            sql.NullString
		userID        sql.NullString
		endorsementID sql.NullString
		createdAt     sql.NullTime
		updatedAt     sql.NullTime
	)
	if err := row.Scan(&id, &userID, &endorsementID, &createdAt, &updatedAt); err != nil {
		return Item{}, fmt.Errorf("scan saved slate item: %w", err)
	}
	return Item{
		ID:            id.String,
		UserID:        userID.String,
		EndorsementID: endorsementID.String,
		CreatedAt:     createdAt.Time.UTC().Round(0),
		UpdatedAt:     updatedAt.Time.UTC().Round(0),
	}, nil
}

var _ Store = (*PostgresStore)(nil)

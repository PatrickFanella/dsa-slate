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
		SELECT id, user_id, record_type, record_id, created_at, updated_at
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

// UpsertForUser saves one public record to a user's slate.
func (s *PostgresStore) UpsertForUser(ctx context.Context, userID string, recordType string, recordID string) (Item, error) {
	userID = strings.TrimSpace(userID)
	recordType = strings.TrimSpace(recordType)
	recordID = strings.TrimSpace(recordID)
	if userID == "" || recordID == "" {
		return Item{}, ErrInvalidReference
	}
	recordType, ok := NormalizeRecordType(recordType)
	if !ok {
		return Item{}, ErrInvalidReference
	}

	var exists bool
	if err := s.pool.QueryRow(ctx, referenceExistsQuery(recordType), recordID).Scan(&exists); err != nil {
		return Item{}, fmt.Errorf("validate saved slate reference: %w", err)
	}
	if !exists {
		return Item{}, ErrInvalidReference
	}

	row := s.pool.QueryRow(ctx, `
		INSERT INTO saved_slate_items (id, user_id, record_type, record_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, record_type, record_id) DO UPDATE SET updated_at = now()
		RETURNING id, user_id, record_type, record_id, created_at, updated_at
	`, itemID(userID, recordType, recordID), userID, recordType, recordID)
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

func referenceExistsQuery(recordType string) string {
	switch recordType {
	case RecordTypeEndorsement:
		return `SELECT EXISTS (SELECT 1 FROM endorsements WHERE id = $1)`
	case RecordTypeMeasurePosition:
		return `SELECT EXISTS (SELECT 1 FROM measure_positions WHERE id = $1)`
	case RecordTypeVoterGuide:
		return `SELECT EXISTS (SELECT 1 FROM voter_guides WHERE id = $1)`
	default:
		return `SELECT false`
	}
}

func itemID(userID string, recordType string, recordID string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(userID) + "\x00" + strings.TrimSpace(recordType) + "\x00" + strings.TrimSpace(recordID)))
	return "ssi_" + hex.EncodeToString(sum[:])
}

type rowScanner interface {
	Scan(...any) error
}

func scanItem(row rowScanner) (Item, error) {
	var (
		id         sql.NullString
		userID     sql.NullString
		recordType sql.NullString
		recordID   sql.NullString
		createdAt  sql.NullTime
		updatedAt  sql.NullTime
	)
	if err := row.Scan(&id, &userID, &recordType, &recordID, &createdAt, &updatedAt); err != nil {
		return Item{}, fmt.Errorf("scan saved slate item: %w", err)
	}
	return Item{
		ID:         id.String,
		UserID:     userID.String,
		RecordType: recordType.String,
		RecordID:   recordID.String,
		CreatedAt:  createdAt.Time.UTC().Round(0),
		UpdatedAt:  updatedAt.Time.UTC().Round(0),
	}, nil
}

var _ Store = (*PostgresStore)(nil)

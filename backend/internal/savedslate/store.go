package savedslate

import (
	"context"
	"errors"
	"time"
)

// ErrInvalidReference reports a saved slate item pointing at a missing public record.
var ErrInvalidReference = errors.New("invalid reference")

// Supported saved-slate record types.
const (
	RecordTypeEndorsement     = "endorsement"
	RecordTypeMeasurePosition = "measure_position"
	RecordTypeVoterGuide      = "voter_guide"
)

// Item is one public record saved by one authenticated user.
type Item struct {
	ID         string    `json:"id"`
	UserID     string    `json:"userId"`
	RecordType string    `json:"recordType"`
	RecordID   string    `json:"recordId"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// Slate is the current user's saved public-record slate.
type Slate struct {
	Items []Item `json:"items"`
}

// Store persists authenticated users' saved slate items.
type Store interface {
	ListForUser(context.Context, string) ([]Item, error)
	UpsertForUser(context.Context, string, string, string) (Item, error)
	DeleteForUser(context.Context, string, string) error
}

// NormalizeRecordType validates and normalizes a saved-slate record type.
func NormalizeRecordType(recordType string) (string, bool) {
	switch recordType {
	case RecordTypeEndorsement, RecordTypeMeasurePosition, RecordTypeVoterGuide:
		return recordType, true
	default:
		return "", false
	}
}

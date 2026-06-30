package savedslate

import (
	"context"
	"errors"
	"time"
)

// ErrInvalidReference reports a saved slate item pointing at a missing public record.
var ErrInvalidReference = errors.New("invalid reference")

// Item is one public endorsement saved by one authenticated user.
type Item struct {
	ID            string    `json:"id"`
	UserID        string    `json:"userId"`
	EndorsementID string    `json:"endorsementId"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// Slate is the current user's saved endorsement slate.
type Slate struct {
	Items []Item `json:"items"`
}

// Store persists authenticated users' saved slate items.
type Store interface {
	ListForUser(context.Context, string) ([]Item, error)
	UpsertForUser(context.Context, string, string) (Item, error)
	DeleteForUser(context.Context, string, string) error
}

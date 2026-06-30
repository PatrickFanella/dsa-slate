package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

const sessionTokenSize = 32

// GenerateSessionToken returns an opaque URL-safe session token.
func GenerateSessionToken() (string, error) {
	buf := make([]byte, sessionTokenSize)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

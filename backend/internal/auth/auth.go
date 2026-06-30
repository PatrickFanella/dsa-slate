// Package auth provides users, opaque sessions, and session-resolution helpers.
package auth

import "strings"

// Role identifies an application user role.
type Role string

const (
	// RoleUser is the default public/private-alpha user role.
	RoleUser Role = "user"
	// RoleEditor can access editorial workflows.
	RoleEditor Role = "editor"
	// RoleAdmin can access administrative workflows.
	RoleAdmin Role = "admin"
)

// SessionCookieName is the HTTP-only cookie name reserved for opaque sessions.
const SessionCookieName = "dsa_slate_session"

var allowedRoles = map[Role]struct{}{
	RoleUser:   {},
	RoleEditor: {},
	RoleAdmin:  {},
}

// ParseRole normalizes and validates a role string.
func ParseRole(value string) (Role, error) {
	role := Role(strings.TrimSpace(value))
	if role == "" {
		return RoleUser, nil
	}
	if _, ok := allowedRoles[role]; !ok {
		return "", roleValidationError()
	}
	return role, nil
}

// ValidateRole reports whether a role string is allowed.
func ValidateRole(value string) error {
	_, err := ParseRole(value)
	return err
}

func roleValidationError() error {
	return validationError("role must be one of user, editor, admin")
}

func validationError(message string) error {
	return &ValidationError{Message: message}
}

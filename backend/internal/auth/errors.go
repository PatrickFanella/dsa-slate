package auth

// ValidationError reports a rejected auth value.
type ValidationError struct {
	Message string
}

// Error implements error.
func (e *ValidationError) Error() string {
	return e.Message
}

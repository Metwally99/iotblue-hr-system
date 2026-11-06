package apperror

import "errors"

var (
	ErrNotFound          = errors.New("resource not found")
	ErrConflict          = errors.New("resource conflict")
	ErrUnauthorized      = errors.New("authentication required")
	ErrForbidden         = errors.New("permission denied")
	ErrInactive          = errors.New("employee is inactive")
	ErrInvalidState      = errors.New("resource is in an invalid state")
	ErrInsufficientLeave = errors.New("insufficient leave balance")
)

type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string { return "request validation failed" }

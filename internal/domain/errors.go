package domain

import "errors"

var (
	// ErrNotFound indicates the resource does not exist or is not owned by the user.
	ErrNotFound = errors.New("not found")
	// ErrValidation indicates invalid input.
	ErrValidation = errors.New("validation")
	// ErrForbidden indicates the user lacks permission for the action.
	ErrForbidden = errors.New("forbidden")
	// ErrConflict indicates the action cannot proceed due to current state.
	ErrConflict = errors.New("conflict")
)

package model

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrConflict       = errors.New("conflict")
	ErrInvalidInput   = errors.New("invalid input")
	ErrInvalidState   = errors.New("invalid state transition")
	ErrImmutable      = errors.New("immutable resource")
	ErrCircularGraph  = errors.New("circular dependency graph")
	ErrMissingLicense = errors.New("missing license declaration")
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e FieldError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func Invalid(field, message string) error {
	return fmt.Errorf("%w: %w", ErrInvalidInput, FieldError{Field: field, Message: message})
}

func IsInvalid(err error) bool   { return errors.Is(err, ErrInvalidInput) }
func IsConflict(err error) bool  { return errors.Is(err, ErrConflict) }
func IsNotFound(err error) bool  { return errors.Is(err, ErrNotFound) }
func IsImmutable(err error) bool { return errors.Is(err, ErrImmutable) }

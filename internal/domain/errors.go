package domain

import (
	"errors"
	"fmt"
)

var (
	ErrInvalid       = errors.New("invalid")
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	if e.Field == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func (e ValidationError) Unwrap() error { return ErrInvalid }

type NotFoundError struct {
	Resource string
	ID       string
}

func (e NotFoundError) Error() string {
	if e.ID == "" {
		return e.Resource + " not found"
	}
	return e.Resource + " not found: " + e.ID
}

func (e NotFoundError) Unwrap() error { return ErrNotFound }

type AlreadyExistsError struct {
	Resource string
	ID       string
}

func (e AlreadyExistsError) Error() string {
	if e.ID == "" {
		return e.Resource + " already exists"
	}
	return e.Resource + " already exists: " + e.ID
}

func (e AlreadyExistsError) Unwrap() error { return ErrAlreadyExists }

func IsInvalid(err error) bool       { return errors.Is(err, ErrInvalid) }
func IsNotFound(err error) bool      { return errors.Is(err, ErrNotFound) }
func IsAlreadyExists(err error) bool { return errors.Is(err, ErrAlreadyExists) }

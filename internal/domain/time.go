package domain

import (
	"strings"
	"time"
)

func UTC(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	return t.UTC()
}

func requireTime(field string, t time.Time) error {
	if t.IsZero() {
		return ValidationError{Field: field, Message: "is required"}
	}
	return nil
}

func requireName(field, value string, max int) error {
	if strings.TrimSpace(value) == "" {
		return ValidationError{Field: field, Message: "is required"}
	}
	if len(value) > max {
		return ValidationError{Field: field, Message: "is too long"}
	}
	return nil
}

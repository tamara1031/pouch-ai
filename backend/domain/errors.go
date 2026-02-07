package domain

import "errors"

var (
	ErrKeyNotFound      = errors.New("key not found")
	ErrKeyExpired       = errors.New("key has expired")
	ErrInvalidKey       = errors.New("invalid API key")
	ErrBudgetExceeded   = errors.New("budget limit exceeded")
	ErrProviderNotFound = errors.New("provider not found")
)

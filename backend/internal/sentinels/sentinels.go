package sentinels

import "errors"

var (
	ErrInvalidEmail    = errors.New("invalid email")
	ErrInvalidPassword = errors.New("invalid password")
	ErrDBExecFailed    = errors.New("db execution failed")
	ErrInvalidCursor   = errors.New("invalid cursor")
	ErrInvalidID       = errors.New("invalid UUID")
)

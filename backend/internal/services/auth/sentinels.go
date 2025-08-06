package auth

import "errors"

var (
	ErrNoAuthHeaderIncluded = errors.New("no authorization header included")
	ErrInvalidCreds         = errors.New("invalid email or password")
	ErrInvalidDeviceInfo    = errors.New("invalid or unknown device information")
)

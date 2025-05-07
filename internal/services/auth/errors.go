package auth

import "errors"

var (
	ErrNoAuthHeaderIncluded = errors.New("no authorization header included")
	ErrInvalidEmail         = errors.New("invalid email")
	ErrInvalidPassword      = errors.New("invalid password")
)

package cursor

import (
	"errors"
)

var (
	ErrCursorInvalidFormat = errors.New("cursor: invalid token format")
	ErrCursorBadSignature  = errors.New("cursor: bad signature")
)

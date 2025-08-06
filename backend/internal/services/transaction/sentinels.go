package transaction

import "errors"

var (
	ErrInvalidDateFormat          = errors.New("invalid date format")
	ErrTxNotFound                 = errors.New("tx not found")
	ErrInvalidPageSizeNonPositive = errors.New("pageSize must be a positive int")
	ErrDateTime                   = errors.New("invalid datetime conversion")
	ErrPageSizeTooLarge           = errors.New("pageSize too large")
	ErrInconsistentToken          = errors.New("inconsistent cursor token")
	ErrEncodingFailed             = errors.New("cursor encoding failed")
)

package transaction

import "errors"

var (
	ErrInvalidDateFormat = errors.New("invalid date format")
	ErrTxNotFound        = errors.New("tx not found")
	ErrInvalidPageSize   = errors.New("pageSize must be a positive int")
)

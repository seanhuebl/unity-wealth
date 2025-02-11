package services

import (
	"github.com/seanhuebl/unity-wealth/internal/interfaces"
)

type TransactionService struct {
	queries *interfaces.Quierier
}

func NewTransactionService(queries *interfaces.Quierier) *TransactionService {
	return &TransactionService{queries: queries}
}

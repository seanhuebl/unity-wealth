package transaction

import "github.com/seanhuebl/unity-wealth/internal/services/transaction"

type Handler struct {
	txSvc *transaction.TransactionService
}

func NewHandler(txSvc *transaction.TransactionService) *Handler {
	return &Handler{
		txSvc: txSvc,
	}
}

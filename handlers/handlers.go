package handlers

import (
	"github.com/seanhuebl/unity-wealth/internal/interfaces"
	"github.com/seanhuebl/unity-wealth/internal/services/auth"
	"github.com/seanhuebl/unity-wealth/internal/services/transaction"
	"github.com/seanhuebl/unity-wealth/internal/services/user"
)

type Handler struct {
	queries            interfaces.Querier
	transactionService *transaction.TransactionService
	authService        *auth.AuthService
	UserService        *user.UserService
}

func NewHandler(queries interfaces.Querier, txnSvc *transaction.TransactionService, authSvc *auth.AuthService, userSvc *user.UserService) *Handler {
	return &Handler{queries: queries, transactionService: txnSvc, authService: authSvc, UserService: userSvc}
}

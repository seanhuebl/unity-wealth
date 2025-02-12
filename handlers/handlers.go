package handlers

import (
	"github.com/seanhuebl/unity-wealth/internal/interfaces"
	"github.com/seanhuebl/unity-wealth/services"
)

type Handler struct {
	queries            interfaces.Querier
	transactionService *services.TransactionService
	authService        *services.AuthService
	userService        *services.UserService
}

func NewHandler(queries interfaces.Querier, txnSvc *services.TransactionService, authSvc *services.AuthService, userSvc *services.UserService) *Handler {
	return &Handler{queries: queries, transactionService: txnSvc, authService: authSvc, userService: userSvc}
}

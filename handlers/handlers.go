package handlers

import (
	"github.com/seanhuebl/unity-wealth/internal/config"
	"github.com/seanhuebl/unity-wealth/services"
)

type Handler struct {
	cfg                *config.ApiConfig
	transactionService *services.TransactionService
}

func NewHandler(cfg *config.ApiConfig, txnSvc *services.TransactionService) *Handler {
	return &Handler{cfg: cfg, transactionService: txnSvc}
}

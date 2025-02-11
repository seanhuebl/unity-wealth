package services

import (
	"github.com/gin-gonic/gin"
	"github.com/seanhuebl/unity-wealth/handlers"
)

func (s *TransactionService) CreateTransaction(ctx *gin.Context, userID string, req handlers.NewTransactionRequest) (*handlers.Transaction, error) {

}

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/helpers"
	"github.com/seanhuebl/unity-wealth/models"
)

func (h *Handler) NewTransaction(ctx *gin.Context) {
	claims, err := helpers.ValidateClaims(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "uuid parsing error",
		})
		return
	}

	var req models.NewTransactionRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	txn, err := h.transactionService.CreateTransaction(ctx, userID.String(), req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create transaction",
		})
	}
	response := h.transactionService.ConvertToResponse(txn)

	ctx.JSON(http.StatusCreated, response)

}

func (h *Handler) GetTransactionsByUserID(ctx *gin.Context) {
	claims, err := helpers.ValidateClaims(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err,
		})
		return
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "uuid parsing error",
		})
		return
	}
	cursorDate := ctx.Query("cursor_date")
	cursorID := ctx.Query("cursor_id")
	pageSize := int64(50)
	transactions, nextCursorDate, nextCursorID, hasMoreData, err := h.transactionService.ListUserTransactions(ctx, userID, &cursorDate, &cursorID, &pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "uanble to get transactions",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"transactions":     transactions,
		"next_cursor_date": nextCursorDate,
		"next_cursor_id":   nextCursorID,
		"has_more_data":    hasMoreData,
	})
}

func (h *Handler) GetTransactionByID(ctx *gin.Context) {
	claims, err := helpers.ValidateClaims(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err,
		})
		return
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "uuid parsing error",
		})
		return
	}
	id := ctx.Param("id")
	txn, err := h.transactionService.GetTransactionByID(ctx, userID.String(), id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "unable to get transaction",
		})
		return
	}
	response := h.transactionService.ConvertToResponse(txn)
	ctx.JSON(http.StatusOK, response)
}

func (h *Handler) UpdateTransaction(ctx *gin.Context) {
	var req models.NewTransactionRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := ctx.Param("id")

	txn, err := h.transactionService.UpdateTransaction(ctx, id, req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update transaction",
		})
	}

	response := h.transactionService.ConvertToResponse(txn)
	ctx.JSON(http.StatusOK, response)
}

func (h *Handler) DeleteTransaction(ctx *gin.Context) {
	claims, err := helpers.ValidateClaims(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err,
		})
		return
	}

	id := ctx.Param("id")

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "uuid parsing error",
		})
		return
	}

	success, err := h.transactionService.DeleteTransaction(ctx, id, userID.String())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "error deleting transaction",
		})
	}
	ctx.JSON(http.StatusOK, success)
}

package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/helpers"
	"github.com/seanhuebl/unity-wealth/internal/database"
)

type Transaction struct {
	ID               string  `json:"id"`
	UserID           string  `json:"user_id"`
	Date             string  `json:"date" binding:"required"`
	Merchant         string  `json:"merchant" binding:"required"`
	Amount           float64 `json:"amount" binding:"required"`
	DetailedCategory int64   `json:"detailed_category" binding:"required"`
}

func (h *Handler) NewTransaction(ctx *gin.Context) {
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

	var req Transaction

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err = time.Parse("2006-01-02", req.Date)
	if err != nil {

		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid date format: %v",
		})
		return
	}

	if err := h.cfg.Queries.CreateTransaction(ctx, database.CreateTransactionParams{
		ID:                 uuid.NewString(),
		UserID:             userID.String(),
		TransactionDate:    req.Date,
		Merchant:           req.Merchant,
		AmountCents:        int64(req.Amount * 100),
		DetailedCategoryID: req.DetailedCategory,
	}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "unable to create transaction",
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"transaction_created": "success",
	})

}

func (h *Handler) UpdateTransaction(ctx *gin.Context) {
	var req Transaction

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := time.Parse("2006-01-02", req.Date)
	if err != nil {

		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid date format: %v",
		})
		return
	}
	id := ctx.Param("id")

	txRow, err := h.cfg.Queries.UpdateTransactionByID(ctx, database.UpdateTransactionByIDParams{
		TransactionDate:    req.Date,
		Merchant:           req.Merchant,
		AmountCents:        int64(req.Amount * 100),
		DetailedCategoryID: req.DetailedCategory,
		UpdatedAt:          sql.NullTime{Time: time.Now(), Valid: true},
		ID:                 id,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "error updating transaction",
		})
		return
	}
	updatedTx := Transaction{
		ID:               txRow.ID,
		Date:             txRow.TransactionDate,
		Merchant:         txRow.Merchant,
		Amount:           float64(txRow.AmountCents / 100),
		DetailedCategory: req.DetailedCategory,
	}
	ctx.JSON(http.StatusOK, updatedTx)
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

	if err = h.cfg.Queries.DeleteTransactionById(ctx, database.DeleteTransactionByIdParams{
		ID:     id,
		UserID: userID.String(),
	}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "error deleting transaction",
		})
	}
	ctx.JSON(http.StatusOK, gin.H{
		"delete_transaction": "success",
	})
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

	if cursorDate == "" || cursorID == "" {

	}
}
func (h *Handler) listUserTransactions(ctx *gin.Context, userID uuid.UUID, cursorDate *string, cursorID *string, pageSize *int64) (transactions []Transaction, nextCursorDate, nextCursorID string, hasMoreData bool, err error) {
	// Get first page of 50 transactions
	firstPageTransactions, err := h.cfg.Queries.GetUserTransactionsFirstPage(ctx, database.GetUserTransactionsFirstPageParams{UserID: userID.String(), Limit: *pageSize + 1})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "error loading first page of transactions",
		})
		return
	}

	if len(firstPageTransactions) > 0 {
		lastTxnRow := firstPageTransactions[len(firstPageTransactions)-1]
	}

	return transactions, nextCursorDate, nextCursorID, hasMoreData, nil
}

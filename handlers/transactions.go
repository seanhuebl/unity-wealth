package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/cache"
	"github.com/seanhuebl/unity-wealth/internal/config"
	"github.com/seanhuebl/unity-wealth/internal/database"
)

type Transaction struct {
	ID               string  `json:"id"`
	Date             string  `json:"date" binding:"required"`
	Merchant         string  `json:"merchant" binding:"required"`
	Amount           float64 `json:"amount" binding:"required"`
	DetailedCategory string  `json:"detailed_category" binding:"required"`
}

func NewTransaction(ctx *gin.Context, cfg *config.ApiConfig) {
	claimsInterface, exists := ctx.Get("claims")
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "unauthorized: no claims found",
		})
		return
	}
	claims, ok := claimsInterface.(*jwt.RegisteredClaims)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "invalid claims format",
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
	detailedCategories, err := cache.GetCachedDetailedCategories(ctx)
	if err != nil {

	}

	if err := cfg.Queries.CreateTransaction(ctx, database.CreateTransactionParams{
		ID:                 uuid.NewString(),
		UserID:             userID.String(),
		TransactionDate:    req.Date,
		Merchant:           req.Merchant,
		AmountCents:        int64(req.Amount * 100),
		DetailedCategoryID: detailedCategoryID,
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

func UpdateTransaction(ctx *gin.Context, cfg *config.ApiConfig) {
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
	detailedCategoryID, err := cfg.Queries.GetDetailedCategoryId(ctx, req.DetailedCategory)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid detailed category",
		})
		return
	}
	id := ctx.Param("id")

	txRow, err := cfg.Queries.UpdateTransactionByID(ctx, database.UpdateTransactionByIDParams{
		TransactionDate:    req.Date,
		Merchant:           req.Merchant,
		AmountCents:        int64(req.Amount * 100),
		DetailedCategoryID: detailedCategoryID,
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

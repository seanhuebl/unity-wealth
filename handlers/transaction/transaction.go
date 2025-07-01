package transaction

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	"github.com/seanhuebl/unity-wealth/internal/helpers"
	"github.com/seanhuebl/unity-wealth/internal/models"
	"github.com/seanhuebl/unity-wealth/internal/services/transaction"
)

func (h *Handler) NewTransaction(c *gin.Context) {
	userID, err := helpers.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"data": gin.H{
				"error": "unauthorized",
			},
		})
		return
	}

	var req models.NewTxRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"data": gin.H{
				"error": "invalid request body",
			},
		})
		return
	}

	txn, err := h.txSvc.CreateTransaction(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"data": gin.H{
				"error": "failed to create transaction",
			},
		})
		return
	}

	response := models.ConvertToResponse(txn)

	c.JSON(http.StatusCreated, gin.H{
		"data": response,
	})

}

func (h *Handler) GetTransactionsByUserID(c *gin.Context) {
	userID, err := helpers.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"data": gin.H{
				"error": "unauthorized",
			},
		})
		return
	}

	ctx := c.Request.Context()

	curString, ok := ctx.Value(constants.CursorKey).(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"data": gin.H{
				"error": "invalid cursor string",
			},
		})
		return
	}

	pageSize, ok := ctx.Value(constants.LimitKey).(int32)
	if !ok || pageSize <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"data": gin.H{
				"error": "invalid page_size; must be > 0",
			},
		})
		return
	}

	transactions, err := h.txSvc.ListUserTransactions(c.Request.Context(), userID, curString, pageSize)
	if err != nil {
		if strings.Contains(err.Error(), "no transactions found") {
			c.JSON(http.StatusOK, gin.H{
				"data": gin.H{
					"error": "transactions not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"data": gin.H{
				"error": "unable to get transactions",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"transactions":   transactions.Transactions,
			"next_cursor":    transactions.NextCursor,
			"has_more_data":  transactions.HasMoreData,
			"clamped":        transactions.Clamped,
			"effective_size": transactions.EffectiveSize,
		},
	})
}

func (h *Handler) GetTransactionByID(c *gin.Context) {
	userID, err := helpers.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"data": gin.H{
				"error": "unauthorized",
			},
		})
		return
	}

	txId, ok := helpers.BindUUIDParam(c, "id")
	if !ok {
		// response is in the helper
		return
	}
	txn, err := h.txSvc.GetTransactionByID(c.Request.Context(), userID, txId)
	if err != nil {
		if errors.Is(err, transaction.ErrTxNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"data": gin.H{
					"error": "not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"data": gin.H{
				"error": "something went wrong",
			},
		})
		return
	}

	response := models.ConvertToResponse(txn)

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

func (h *Handler) UpdateTransaction(c *gin.Context) {
	userID, err := helpers.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"data": gin.H{
				"error": "unauthorized",
			},
		})
		return
	}

	var req models.NewTxRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"data": gin.H{
				"error": "invalid request body",
			},
		})
		return
	}

	txId, ok := helpers.BindUUIDParam(c, "id")
	if !ok {
		// response is in the helper
		return
	}

	txn, err := h.txSvc.UpdateTransaction(c.Request.Context(), txId, userID, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"data": gin.H{
					"error": "not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"data": gin.H{
				"error": "failed to update transaction",
			},
		})
		return
	}

	response := models.ConvertToResponse(txn)

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

func (h *Handler) DeleteTransaction(c *gin.Context) {
	userID, err := helpers.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"data": gin.H{
				"error": "unauthorized",
			},
		})
		return
	}

	txId, ok := helpers.BindUUIDParam(c, "id")
	if !ok {
		// response is in the helper
		return
	}
	err = h.txSvc.DeleteTransaction(c.Request.Context(), txId, userID)

	if err != nil {
		if errors.Is(err, transaction.ErrTxNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"data": gin.H{
					"error": "not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"data": gin.H{
				"error": "something went wrong",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"transaction_deleted": "success",
		},
	})
}

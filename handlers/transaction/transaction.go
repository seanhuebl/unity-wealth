package transaction

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	"github.com/seanhuebl/unity-wealth/internal/helpers"
	"github.com/seanhuebl/unity-wealth/internal/models"
)

func (h *Handler) NewTransaction(ctx *gin.Context) {
	userID, err := helpers.GetUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"data": gin.H{
				"error": "unauthorized",
			},
		})
		return
	}

	var req models.NewTxRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data": gin.H{
				"error": "invalid request body",
			},
		})
		return
	}

	txn, err := h.txSvc.CreateTransaction(ctx.Request.Context(), userID.String(), req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"data": gin.H{
				"error": "failed to create transaction",
			},
		})
		return
	}

	response := models.ConvertToResponse(txn)

	ctx.JSON(http.StatusCreated, gin.H{
		"data": response,
	})

}

func (h *Handler) GetTransactionsByUserID(ctx *gin.Context) {
	userID, err := helpers.GetUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"data": gin.H{
				"error": "unauthorized",
			},
		})
		return
	}

	cursorDateVal, exists := ctx.Get(string(constants.CursorDateKey))
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data": gin.H{
				"error": "cursor date key not set in context",
			},
		})
		return
	}
	cursorDateStr, ok := cursorDateVal.(string)
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data": gin.H{
				"error": "invalid cursor date",
			},
		})
		return
	}
	cursorIDVal, exists := ctx.Get(string(constants.CursorIDKey))
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data": gin.H{
				"error": "cursor ID key not set in context",
			},
		})
		return
	}
	cursorIDStr, ok := cursorIDVal.(string)
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data": gin.H{
				"error": "invalid cursor ID",
			},
		})
		return
	}
	pageSizeVal, exists := ctx.Get(string(constants.PageSizeKey))
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data": gin.H{
				"error": "page_size not provided",
			},
		})
		return
	}
	pageSizeInt, ok := pageSizeVal.(int)
	if !ok || pageSizeInt <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data": gin.H{
				"error": "invalid page_size; must be > 0",
			},
		})
		return
	}
	pageSize := int64(pageSizeInt)

	var cursorDatePtr *string
	if cursorDateStr != "" {
		cursorDatePtr = &cursorDateStr
	}

	var cursorIDPtr *string
	if cursorIDStr != "" {
		cursorIDPtr = &cursorIDStr
	}

	transactions, nextCursorDate, nextCursorID, hasMoreData, err :=
		h.txSvc.ListUserTransactions(ctx.Request.Context(), userID, cursorDatePtr, cursorIDPtr, pageSize)
	if err != nil {
		if strings.Contains(err.Error(), "no transactions found") {
			ctx.JSON(http.StatusOK, gin.H{
				"data": gin.H{
					"error": "transactions not found",
				},
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"data": gin.H{
				"error": "unable to get transactions",
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"transactions":     transactions,
			"next_cursor_date": nextCursorDate,
			"next_cursor_id":   nextCursorID,
			"has_more_data":    hasMoreData,
		},
	})
}

func (h *Handler) GetTransactionByID(ctx *gin.Context) {
	userID, err := helpers.GetUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"data": gin.H{
				"error": "unauthorized",
			},
		})
		return
	}

	txId, ok := helpers.BindUUIDParam(ctx, "id")
	if !ok {
		// response is in the helper
		return
	}
	txn, err := h.txSvc.GetTransactionByID(ctx.Request.Context(), userID.String(), txId.String())
	if err != nil {
		if strings.Contains(err.Error(), "transaction not found") {
			ctx.JSON(http.StatusNotFound, gin.H{
				"data": gin.H{
					"error": "not found",
				},
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"data": gin.H{
				"error": "unable to get transaction",
			},
		})
		return
	}

	response := models.ConvertToResponse(txn)

	ctx.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

func (h *Handler) UpdateTransaction(ctx *gin.Context) {
	userID, err := helpers.GetUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"data": gin.H{
				"error": "unauthorized",
			},
		})
		return
	}

	var req models.NewTxRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data": gin.H{
				"error": "invalid request body",
			},
		})
		return
	}

	txId, ok := helpers.BindUUIDParam(ctx, "id")
	if !ok {
		// response is in the helper
		return
	}

	txn, err := h.txSvc.UpdateTransaction(ctx.Request.Context(), txId.String(), userID.String(), req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			ctx.JSON(http.StatusNotFound, gin.H{
				"data": gin.H{
					"error": "not found",
				},
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"data": gin.H{
				"error": "failed to update transaction",
			},
		})
		return
	}

	response := models.ConvertToResponse(txn)

	ctx.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

func (h *Handler) DeleteTransaction(ctx *gin.Context) {
	userID, err := helpers.GetUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"data": gin.H{
				"error": "unauthorized",
			},
		})
		return
	}

	txId, ok := helpers.BindUUIDParam(ctx, "id")
	if !ok {
		// response is in the helper
		return
	}

	err = h.txSvc.DeleteTransaction(ctx.Request.Context(), txId.String(), userID.String())
	if err != nil {
		if strings.Contains(err.Error(), "transaction not found") {
			ctx.JSON(http.StatusNotFound, gin.H{
				"data": gin.H{
					"error": "not found",
				},
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"data": gin.H{
				"error": "error deleting transaction",
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"transaction_deleted": "success",
		},
	})
}

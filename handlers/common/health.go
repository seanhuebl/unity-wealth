package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Health(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

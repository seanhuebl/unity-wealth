package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func health(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

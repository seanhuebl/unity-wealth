package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/seanhuebl/unity-wealth/models"
)

// POST
func (h *Handler) AddUser(ctx *gin.Context) {
	var input models.SignUpInput

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}
	if err := h.UserService.SignUp(ctx, input); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create user",
		})
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"data": gin.H{
			"message": "sign up successful!",
			"email":   input.Email,
		},
	})
}

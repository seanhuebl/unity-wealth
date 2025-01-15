package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/seanhuebl/unity-wealth/internal/auth"
	"github.com/seanhuebl/unity-wealth/internal/database"
)

type SignUpInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// POST
func addUser(ctx *gin.Context, cfg *ApiConfig) {
	var input SignUpInput

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	hashedPW, err := auth.HashPassword(input.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	cfg.Queries.CreateUser(context.Background(), database.CreateUserParams{
		Email:          input.Email,
		HashedPassword: hashedPW,
	})

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Sign up successful!",
		"email":   input.Email,
	})
}

package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/auth"
)

type LoginInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Login(ctx *gin.Context, cfg *ApiConfig) {
	var input LoginInput

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	user, err := cfg.Queries.GetUserByEmail(ctx, input.Email)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = auth.CheckPasswordHash(input.Password, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}
	// Generate and pass across JWT and refresh token
	JWT, err := auth.MakeJWT(user.ID.(uuid.UUID), cfg.TokenSecret, time.Minute*15)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	cookie := http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		Domain:   "localhost", // Use 'localhost' for local testing
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   false, // Disable 'Secure' for HTTP testing
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(ctx.Writer, &cookie)

	

	ctx.JSON(http.StatusOK, gin.H{
		"token": JWT,
	})
}

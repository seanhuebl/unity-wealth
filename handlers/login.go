package handlers

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/seanhuebl/unity-wealth/models"
)

func (h *Handler) Login(ctx *gin.Context) {
	var input models.LoginInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}

	loginResp, err := h.authService.Login(ctx.Request.Context(), input)
	if err != nil {
		// Return a generic error message for the client.
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "login failed"})
		return
	}

	// Set the refresh token cookie (HTTP-specific).
	SetRefreshTokenCookie(ctx, loginResp.RefreshToken)

	ctx.JSON(http.StatusOK, gin.H{"data": gin.H{"token": loginResp.JWT}})
}

// Helpers
func SetRefreshTokenCookie(ctx *gin.Context, refreshToken string) {
	isProduction := os.Getenv("ENV") == "prod"
	cookieDomain := os.Getenv("COOKIE_DOMAIN")
	if cookieDomain == "" {
		cookieDomain = "localhost"
	}

	cookie := http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		Domain:   cookieDomain, // Use 'localhost' for local testing
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   isProduction, // Disable 'Secure' for HTTP testing
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(ctx.Writer, &cookie)
}

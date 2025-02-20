package helpers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func ValidateClaims(ctx *gin.Context) (*jwt.RegisteredClaims, error) {
	claimsInterface, exists := ctx.Get("claims")
	if !exists {
		return nil, fmt.Errorf("unauthorized: no claims found")
	}

	claims, ok := claimsInterface.(*jwt.RegisteredClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims format")
	}
	return claims, nil
}

package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/helpers"
)

type contextKey string

const (
	userIDKey contextKey = "userID"
	claimsKey contextKey = "claims"
)

func (m *Middleware) ClaimsAuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims, err := helpers.ValidateClaims(ctx)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			return
		}

		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid user ID in token",
			})
			return
		}

		ctx.Set(string(claimsKey), claims)
		ctx.Set(string(userIDKey), userID)

		ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), claimsKey, claims))
		ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), userIDKey, userID))

		ctx.Next()
	}
}

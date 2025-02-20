package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	"github.com/seanhuebl/unity-wealth/internal/helpers"
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

		ctx.Set(string(constants.ClaimsKey), claims)
		ctx.Set(string(constants.UserIDKey), userID)

		ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), constants.ClaimsKey, claims))
		ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), constants.UserIDKey, userID))
		// Store the request in the standard context.
		ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), constants.RequestKey, ctx.Request))
		ctx.Next()
	}
}

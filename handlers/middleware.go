package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/seanhuebl/unity-wealth/internal/auth"
	"github.com/seanhuebl/unity-wealth/internal/config"
)

func UserAuthMiddleware(cfg *config.ApiConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authSvc := auth.NewAuthService()

		token, err := authSvc.GetBearerToken(ctx.Request.Header)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}
		claims, err := authSvc.ValidateJWT(token, cfg.TokenSecret)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
			})
			return
		}

		ctx.Set("claims", claims)
		ctx.Next()
	}
}

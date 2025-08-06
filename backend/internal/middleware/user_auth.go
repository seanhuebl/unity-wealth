package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (m *Middleware) UserAuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		token, err := m.tokenExtractor.GetBearerToken(ctx.Request.Header)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}
		claims, err := m.tokenGen.ValidateJWT(token)
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

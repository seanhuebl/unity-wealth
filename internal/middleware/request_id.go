package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/constants"
)

func (m *Middleware) RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("X-Request-ID")

		id, err := uuid.Parse(raw)
		if err != nil {
			id = uuid.New()
		}

		c.Writer.Header().Set("X-Request-ID", id.String())
		ctx := context.WithValue(c.Request.Context(), constants.RequestIDKey, id)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/seanhuebl/unity-wealth/internal/constants"
)

func (m *Middleware) Paginate() gin.HandlerFunc {
	return func(c *gin.Context) {
		cur := c.Query(string(constants.CursorKey))
		limit := c.DefaultQuery(string(constants.LimitKey), string(constants.MaxPageSize))

		ctx := context.WithValue(c.Request.Context(), constants.CursorKey, cur)
		ctx = context.WithValue(ctx, constants.LimitKey, limit)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

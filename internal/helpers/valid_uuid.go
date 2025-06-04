package helpers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func BindUUIDParam(c *gin.Context, paramName string) (uuid.UUID, bool) {
	s := c.Param(paramName)
	id, err := uuid.Parse(s)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"data": gin.H{
				"error": fmt.Sprintf("invalid %s", paramName),
			},
		})
		return uuid.Nil, false
	}

	return id, true
}

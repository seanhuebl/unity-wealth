package helpers

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/constants"
)

// GetUserID retrieves the user ID stored in the Gin context or standard context.
func GetUserID(ctx context.Context) (uuid.UUID, error) {
	// First, try to get from Gin context.
	if ginCtx, ok := ctx.(*gin.Context); ok {
		if uid, exists := ginCtx.Get(string(constants.UserIDKey)); exists && uid != nil && uid != uuid.Nil {
			if userID, ok := uid.(uuid.UUID); ok {
				return userID, nil
			}
			return uuid.Nil, fmt.Errorf("user ID is not UUID")
		}
	}
	if uid := ctx.Value(constants.UserIDKey); uid != nil && uid != uuid.Nil {
		if userID, ok := uid.(uuid.UUID); ok {
			return userID, nil
		} else {
			return uuid.Nil, fmt.Errorf("user ID is not UUID")
		}
	}
	return uuid.Nil, fmt.Errorf("user ID not found in context")
}

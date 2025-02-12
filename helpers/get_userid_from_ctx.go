package helpers

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type contextKey string

const userIDKey = contextKey("userID")

// GetUserID retrieves the user ID stored in the Gin context or standard context.
func GetUserID(ctx context.Context) (uuid.UUID, error) {
	// First, try to get from Gin context.
	if uid := ctx.Value(string(userIDKey)); uid != nil {
		if userID, ok := uid.(uuid.UUID); ok {
			return userID, nil
		}
	}
	return uuid.Nil, fmt.Errorf("user ID not found in context")
}

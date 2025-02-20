package helpers

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/constants"
)

// GetUserID retrieves the user ID stored in the Gin context or standard context.
func GetUserID(ctx context.Context) (uuid.UUID, error) {
	// First, try to get from Gin context.
	if uid := ctx.Value(string(constants.UserIDKey)); uid != nil {
		if userID, ok := uid.(uuid.UUID); ok {
			return userID, nil
		}
	}
	return uuid.Nil, fmt.Errorf("user ID not found in context")
}

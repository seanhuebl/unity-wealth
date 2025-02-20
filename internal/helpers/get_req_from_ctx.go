package helpers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/seanhuebl/unity-wealth/internal/constants"
)

func GetRequestFromContext(ctx context.Context) (*http.Request, error) {
	if req, ok := ctx.Value(constants.RequestKey).(*http.Request); ok {
		return req, nil
	}
	return nil, fmt.Errorf("http request not found in context")
}

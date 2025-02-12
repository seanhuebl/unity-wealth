package helpers

import (
	"context"
	"fmt"
	"net/http"
)

const requestKey = contextKey("httpRequest")

func GetRequestFromContext(ctx context.Context) (*http.Request, error) {
	if req, ok := ctx.Value(requestKey).(*http.Request); ok {
		return req, nil
	}
	return nil, fmt.Errorf("http request not found in context")
}

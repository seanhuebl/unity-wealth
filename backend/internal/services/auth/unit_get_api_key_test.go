package auth_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/seanhuebl/unity-wealth/internal/services/auth"
)

func TestGetAPIKey(t *testing.T) {
	t.Parallel()
	tokenExtractor := auth.NewRealTokenExtractor()

	tests := map[string]struct {
		input         http.Header
		expectedValue string
	}{
		"simple":                 {input: http.Header{"Authorization": []string{"ApiKey 1234"}}, expectedValue: "1234"},
		"wrong auth header":      {input: http.Header{"Authorization": []string{"Bearer 1234"}}, expectedValue: "malformed authorization header"},
		"incomplete auth header": {input: http.Header{"Authorization": []string{"ApiKey "}}, expectedValue: "malformed authorization header"},
		"no auth header":         {input: http.Header{"Authorization": []string{""}}, expectedValue: fmt.Sprint(auth.ErrNoAuthHeaderIncluded)},
	}

	for test, tt := range tests {
		t.Run(test, func(t *testing.T) {
			t.Parallel()

			receivedValue, err := tokenExtractor.GetAPIKey(tt.input)
			var diff string
			if err != nil {
				diff = cmp.Diff(tt.expectedValue, fmt.Sprint(err))
			} else {
				diff = cmp.Diff(tt.expectedValue, receivedValue)
			}
			if diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

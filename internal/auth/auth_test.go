package auth

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGetAPIKey(t *testing.T) {
	tests := map[string]struct {
		input         http.Header
		expectedValue string
	}{
		"simple":                 {input: http.Header{"Authorization": []string{"ApiKey 1234"}}, expectedValue: "1234"},
		"wrong auth header":      {input: http.Header{"Authorization": []string{"Bearer 1234"}}, expectedValue: "malformed authorization header"},
		"incomplete auth header": {input: http.Header{"Authorization": []string{"ApiKey "}}, expectedValue: "malformed authorization header"},
		"no auth header":         {input: http.Header{"Authorization": []string{""}}, expectedValue: fmt.Sprint(ErrNoAuthHeaderIncluded)},
	}

	for test, tc := range tests {
		t.Run(test, func(t *testing.T) {
			receivedValue, err := GetAPIKey(tc.input)
			var diff string
			if err != nil {
				diff = cmp.Diff(tc.expectedValue, fmt.Sprint(err))
			} else {
				diff = cmp.Diff(tc.expectedValue, receivedValue)
			}
			if diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
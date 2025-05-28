package testhelpers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
)

func CheckHTTPResponse(t *testing.T, w *httptest.ResponseRecorder, expErr string, expStatus int, expResp, actualResponse map[string]interface{}) {

	if diff := cmp.Diff(expStatus, w.Code); diff != "" {
		t.Errorf("status code mismatch (-want, +got)\n%s", diff)
	}

	data := actualResponse["data"].(map[string]interface{})

	if expErr != "" {
		require.Contains(t, data["error"].(string), expErr)
	}

	ignoreTokens := cmpopts.IgnoreMapEntries(
		func(key, _ interface{}) bool {
			k, ok := key.(string)
			return ok && k == "token"
		},
	)
	if diff := cmp.Diff(expResp, actualResponse, ignoreTokens); diff != "" {
		t.Errorf("response mismatch (-want, +got)\n%s", diff)
	}

}

func ProcessResponse(w *httptest.ResponseRecorder, t *testing.T) map[string]interface{} {
	var actualResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &actualResponse)
	require.NoError(t, err)
	// Since we are using maps the dc is a float64 which doesn't match the struct
	// so we need to convert to int
	actualResponse = ConvertResponseFloatToInt(actualResponse)
	return actualResponse
}

package testhelpers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/seanhuebl/unity-wealth/internal/testinterfaces"
	"github.com/stretchr/testify/require"
)

func ConvertResponseFloatToInt(actualResponse map[string]interface{}) map[string]interface{} {

	data, ok := actualResponse["data"].(map[string]interface{})
	if !ok {
		return actualResponse
	}
	if dc, ok := data["detailed_category"].(float64); ok {
		data["detailed_category"] = int(dc)
		actualResponse["data"] = data
		return actualResponse
	}
	if transactions, ok := data["transactions"].([]interface{}); ok {
		for _, item := range transactions {
			if tx, ok := item.(map[string]interface{}); ok {
				if dc, ok := tx["detailed_category"].(float64); ok {
					tx["detailed_category"] = int(dc)
				}
			}
		}
	}
	actualResponse["data"] = data
	return actualResponse
}

func CheckTxHTTPResponse[T testinterfaces.BaseAccess](t *testing.T, w *httptest.ResponseRecorder, tc T, actualResponse map[string]interface{}) {
	tcBase := tc.BaseAccess()
	if tcBase.ExpectedError != "" {
		data := actualResponse["data"].(map[string]interface{})
		require.Contains(t, data["error"].(string), tcBase.ExpectedError)
	} else {
		if diff := cmp.Diff(tcBase.ExpectedResponse, actualResponse); diff != "" {
			t.Errorf("response mismatch (-want, +got)\n%s", diff)
		}
	}
	if diff := cmp.Diff(tcBase.ExpectedStatusCode, w.Code); diff != "" {
		t.Errorf("status code mismatch (-want, +got)\n%s", diff)
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

package testhelpers

import (
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/seanhuebl/unity-wealth/internal/testmodels"
	"github.com/stretchr/testify/require"
)

func GetEmailFromBody(reqBody string) string {
	re := regexp.MustCompile(`"email"\s*:\s*"([^"]+)"`)
	matches := re.FindStringSubmatch(reqBody)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func GetPasswordFromBody(reqBody string) string {
	re := regexp.MustCompile(`"password"\s*:\s*"([^"]+)"`)
	matches := re.FindStringSubmatch(reqBody)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func CheckUserHTTPResponse(t *testing.T, w *httptest.ResponseRecorder, tc testmodels.SignUpTest, actualResponse map[string]interface{}) {

	if tc.WantErrSubstr != "" {
		require.Contains(t, actualResponse["error"].(string), tc.WantErrSubstr)
	} else {
		if diff := cmp.Diff(tc.ExpectedResponse, actualResponse); diff != "" {
			t.Errorf("response mismatch (-want, +got)\n%s", diff)
		}
	}
	if diff := cmp.Diff(tc.ExpectedStatusCode, w.Code); diff != "" {
		t.Errorf("status code mismatch (-want, +got)\n%s", diff)
	}
}

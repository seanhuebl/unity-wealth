package transaction_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	"github.com/seanhuebl/unity-wealth/internal/testhelpers"

	"github.com/seanhuebl/unity-wealth/internal/testmodels"
)

func TestIntegrationNewTx(t *testing.T) {
	tests := []testmodels.CreateTxTestCase{
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
				Name:               "success",
				UserID:             uuid.New(),
				ExpectedStatusCode: http.StatusCreated,
				ExpectedResponse: map[string]interface{}{
					"data": map[string]interface{}{
						"date":              "2025-03-05",
						"merchant":          "costco",
						"amount":            125.98,
						"detailed_category": 40,
					},
				},
			},
			ReqBody: `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
		},
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
				Name:               "unauthorized userID: ID is not UUID ",
				ExpectedStatusCode: http.StatusUnauthorized,
				ExpectedResponse: map[string]interface{}{
					"error": "unauthorized",
				},
			},
			ReqBody: `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
		},
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
				Name:               "unauthorized userID: context with invalid userID",
				UserID:             uuid.Nil,
				ExpectedStatusCode: http.StatusUnauthorized,
				ExpectedResponse: map[string]interface{}{
					"error": "unauthorized",
				},
			},
			ReqBody: `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
		},
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
				Name:               "invalid request body",
				UserID:             uuid.New(),
				ExpectedStatusCode: http.StatusBadRequest,
				ExpectedResponse: map[string]interface{}{
					"error": "invalid request body",
				},
			},
			ReqBody: `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40`,
		},
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
				Name:               "failed to create transaction",
				UserID:             uuid.New(),
				ExpectedStatusCode: http.StatusInternalServerError,
				ExpectedResponse: map[string]interface{}{
					"error": "failed to create transaction",
				},
			},
			ReqBody: `{"date": "01/01/99", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			env := testhelpers.SetupTestEnv(t)
			defer env.Db.Close()

			testhelpers.SeedTestUser(t, env.UserQ, tc.UserID)
			testhelpers.SeedTestCategories(t, env.Db)
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/transactions", bytes.NewBufferString(tc.ReqBody))
			req.Header.Set("Content-Type", "application/json")

			if strings.Contains(tc.Name, "ID is not UUID") {
				req = req.WithContext(context.WithValue(req.Context(), constants.UserIDKey, "not-a-uuid"))
			} else {
				req = req.WithContext(context.WithValue(req.Context(), constants.UserIDKey, tc.UserID))
			}

			env.Router.POST("/transactions", env.Handler.NewTransaction)
			env.Router.ServeHTTP(w, req)

			actualResponse := testhelpers.ProcessResponse(w, t)
			testhelpers.CheckTxHTTPResponse(t, w, tc, actualResponse)
		})
	}

}

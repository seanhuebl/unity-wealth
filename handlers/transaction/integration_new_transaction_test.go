package transaction_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/seanhuebl/unity-wealth/internal/testfixtures"
	"github.com/seanhuebl/unity-wealth/internal/testhelpers"

	"github.com/seanhuebl/unity-wealth/internal/testmodels"
)

func TestIntegrationNewTx(t *testing.T) {
	t.Parallel()
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
			BaseHTTPTestCase: testfixtures.InvalidUserID,
			ReqBody:          `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
		},
		{
			BaseHTTPTestCase: testfixtures.NilUserID,
			ReqBody:          `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
		},
		{
			BaseHTTPTestCase: testfixtures.InvalidReqBody,
			ReqBody:          `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40`,
		},

		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
				Name:               "failed to create transaction: invalid date format",
				UserID:             uuid.New(),
				ExpectedStatusCode: http.StatusInternalServerError,
				ExpectedResponse: map[string]interface{}{
					"data": map[string]interface{}{
						"error": "failed to create transaction",
					},
				},
			},
			ReqBody: `{"date": "01/01/99", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			env := testhelpers.SetupTestEnv(t)
			t.Cleanup(func() { env.Db.Close() })
			testhelpers.SeedTestUser(t, env.UserQ, tc.UserID, false)
			testhelpers.SeedTestCategories(t, env.Db)
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/transactions", bytes.NewBufferString(tc.ReqBody))
			req.Header.Set("Content-Type", "application/json")
			env.Router.Use(func(c *gin.Context) {
				testhelpers.CheckForUserIDIssues(tc.Name, tc.UserID, c)
				c.Next()
			})

			env.Router.POST("/transactions", env.Middleware.RequestID(), env.Handlers.TxHandler.NewTransaction)
			env.Router.ServeHTTP(w, req)
			actualResponse := testhelpers.ProcessResponse(w, t)
			testhelpers.CheckHTTPResponse(t, w, tc.ExpectedError, tc.ExpectedStatusCode, tc.ExpectedResponse, actualResponse)
		})
	}

}

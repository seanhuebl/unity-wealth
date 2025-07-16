package transaction_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/testfixtures"
	"github.com/seanhuebl/unity-wealth/internal/testhelpers"
	"github.com/seanhuebl/unity-wealth/internal/testmodels"
)

func TestIntegrationGetTransactionByID(t *testing.T) {
	tests := []testmodels.GetTxTestCase{
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
				Name:               "success",
				UserID:             uuid.New(),
				ExpectedStatusCode: http.StatusOK,
				ExpectedResponse: map[string]interface{}{
					"data": map[string]interface{}{
						"date":              "2025-03-05",
						"merchant":          "costco",
						"amount":            125.98,
						"detailed_category": 40,
					},
				},
			},
			TxID: uuid.New(),
		},
		{
			BaseHTTPTestCase: testfixtures.NilUserID,
			TxID:             uuid.New(),
		},
		{
			BaseHTTPTestCase: testfixtures.InvalidUserID,
			TxID:             uuid.New(),
		},
		{
			BaseHTTPTestCase: testfixtures.InvalidTxID,
			TxIDRaw:             "INVALID",
		},
		{
			BaseHTTPTestCase: testfixtures.NotFound,
			TxID:             uuid.New(),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			env := testhelpers.SetupTestEnv(t)
			defer env.Db.Close()

			if tc.TxIDRaw != "" {
				testhelpers.SeedTestUser(t, env.UserQ, tc.UserID, false)
				testhelpers.SeedTestCategories(t, env.Db)
				testhelpers.IsTxFound(t, tc.BaseHTTPTestCase, tc.TxID, env)
			}
			w := httptest.NewRecorder()

			req := httptest.NewRequest("GET", fmt.Sprintf("/transactions/%v", tc.TxID), nil)

			if tc.TxIDRaw == "" {
				c, _ := gin.CreateTestContext(w)
				c.Request = req
				c.Params = gin.Params{{Key: "id", Value: ""}}
				testhelpers.CheckForUserIDIssues(tc.Name, tc.UserID, c)
				env.Handlers.TxHandler.GetTransactionByID(c)
			} else {
				env.Router.GET("/transactions/:id", func(c *gin.Context) {
					testhelpers.CheckForUserIDIssues(tc.Name, tc.UserID, c)
					env.Handlers.TxHandler.GetTransactionByID(c)
				})
				env.Router.ServeHTTP(w, req)
			}
			actualResponse := testhelpers.ProcessResponse(w, t)
			testhelpers.CheckHTTPResponse(t, w, tc.ExpectedError, tc.ExpectedStatusCode, tc.ExpectedResponse, actualResponse)
		})
	}
}

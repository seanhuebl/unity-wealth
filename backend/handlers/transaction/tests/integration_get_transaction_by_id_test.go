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
						"date":              "2025-03-05T00:00:00Z",
						"merchant":          "costco",
						"amount":            125.98,
						"detailed_category": 40,
					},
				},
			},
			TxID: uuid.New(),
		},
		{
			BaseHTTPTestCase: testfixtures.NilUserID(),
			TxID:             uuid.New(),
		},
		{
			BaseHTTPTestCase: testfixtures.InvalidUserID(),
			TxID:             uuid.New(),
		},
		{
			BaseHTTPTestCase: testfixtures.NotFound(),
			TxID:             uuid.New(),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			env := testhelpers.SetupTestEnv(t)
			t.Cleanup(func() { env.Db.Close() })

			testhelpers.SeedTestUser(t, env.UserQ, tc.UserID, false)
			testhelpers.IsTxFound(t, tc.BaseHTTPTestCase, tc.TxID, env)
			w := httptest.NewRecorder()

			req := httptest.NewRequest("GET", fmt.Sprintf("/transactions/%v", tc.TxID), nil)
			env.Router.Use(func(c *gin.Context) {
				c.Params = gin.Params{{Key: "id", Value: tc.TxID.String()}}
				testhelpers.CheckForUserIDIssues(tc.Name, tc.UserID, c)
				c.Next()
			})
			env.Router.NoRoute(func(c *gin.Context) {
				c.JSON(http.StatusNotFound, gin.H{
					"data": gin.H{"error": "not found"},
				})
			})
			env.Router.GET("/transactions/:id", env.Middleware.RequestID(), env.Handlers.TxHandler.GetTransactionByID)
			env.Router.ServeHTTP(w, req)
			actualResponse := testhelpers.ProcessResponse(w, t)
			testhelpers.CheckHTTPResponse(t, w, tc.ExpectedError, tc.ExpectedStatusCode, tc.ExpectedResponse, actualResponse)
		})
	}
	tc := testmodels.GetTxTestCase{
		BaseHTTPTestCase: testfixtures.InvalidTxID(),
		TxIDRaw:          "INVALID",
	}

	t.Run(tc.Name, func(t *testing.T) {
		t.Parallel()
		env := testhelpers.SetupTestEnv(t)
		t.Cleanup(func() { env.Db.Close() })
		w := httptest.NewRecorder()

		req := httptest.NewRequest("GET", fmt.Sprintf("/transactions/%v", tc.TxIDRaw), nil)
		env.Router.Use(func(c *gin.Context) {
			c.Params = gin.Params{{Key: "id", Value: tc.TxIDRaw}}
			testhelpers.CheckForUserIDIssues(tc.Name, tc.UserID, c)
			c.Next()
		})
		env.Router.NoRoute(func(c *gin.Context) {
			c.JSON(http.StatusNotFound, gin.H{
				"data": gin.H{"error": "not found"},
			})
		})
		env.Router.GET("/transactions/:id", env.Middleware.RequestID(), env.Middleware.Paginate(), env.Handlers.TxHandler.GetTransactionByID)
		env.Router.ServeHTTP(w, req)
		actualResponse := testhelpers.ProcessResponse(w, t)
		testhelpers.CheckHTTPResponse(t, w, tc.ExpectedError, tc.ExpectedStatusCode, tc.ExpectedResponse, actualResponse)
	})
}

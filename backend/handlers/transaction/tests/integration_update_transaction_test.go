package transaction_test

import (
	"bytes"
	"errors"
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

func TestIntegrationUpdateTx(t *testing.T) {
	t.Parallel()
	txID := uuid.New()
	tests := []testmodels.UpdateTxTestCase{
		{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
					Name:               "success",
					UserID:             uuid.New(),
					ExpectedStatusCode: http.StatusOK,
					ExpectedResponse: map[string]interface{}{
						"data": map[string]interface{}{
							"date":              "2025-03-05T00:00:00Z",
							"merchant":          "costco",
							"amount":            400.00,
							"detailed_category": 40,
						},
					},
				},
				TxID:    txID,
				TxIDRaw: txID.String(),
			},
			ReqBody: `{"date": "2025-03-05", "merchant": "costco", "amount": 400.00, "detailed_category": 40}`,
		},
		{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testfixtures.InvalidReqBody(),
				TxID:             txID,
				TxIDRaw:          txID.String(),
			},
			ReqBody: `{"date": "2025-03-05", "merchant": "costco", "amount": 400.00, "detailed_category": 40`,
		},
		{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testfixtures.NilUserID(),
				TxID:             txID,
				TxIDRaw:          txID.String(),
			},
		},
		{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testfixtures.InvalidUserID(),
				TxID:             txID,
				TxIDRaw:          txID.String(),
			},
		},
		{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testfixtures.NotFound(),
				TxID:             txID,
				TxIDRaw:          txID.String(),
			},
			ReqBody: `{"date": "2025-03-05", "merchant": "costco", "amount": 400.00, "detailed_category": 40}`,
		},
		{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
					Name:               "error updating tx",
					UserID:             uuid.New(),
					ExpectedStatusCode: http.StatusInternalServerError,
					ExpectedResponse: map[string]interface{}{
						"data": map[string]interface{}{
							"error": "failed to update transaction",
						},
					},
				},
				TxID:  uuid.New(),
				TxErr: errors.New("failed to update transaction"),
			},
			ReqBody: `{"date": "1/1/1994", "merchant": "costco", "amount": 400.00, "detailed_category": 40}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			env := testhelpers.SetupTestEnv(t)
			defer env.Db.Close()

			if tc.TxIDRaw != "" {
				testhelpers.SeedTestUser(t, env.UserQ, tc.UserID, false)
				testhelpers.IsTxFound(t, tc.BaseHTTPTestCase, tc.TxID, env)
			}
			w := httptest.NewRecorder()

			req := httptest.NewRequest("POST", fmt.Sprintf("/transactions/%v", tc.TxID), bytes.NewBufferString(tc.ReqBody))
			req.Header.Set("Content-Type", "application/json")
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

			env.Router.POST("/transactions/:id", env.Middleware.RequestID(), env.Handlers.TxHandler.UpdateTransaction)

			env.Router.ServeHTTP(w, req)
			actualResponse := testhelpers.ProcessResponse(w, t)
			testhelpers.CheckHTTPResponse(t, w, tc.ExpectedError, tc.ExpectedStatusCode, tc.ExpectedResponse, actualResponse)
		})
	}
	tc := testmodels.UpdateTxTestCase{
		GetTxTestCase: testmodels.GetTxTestCase{
			BaseHTTPTestCase: testfixtures.InvalidTxID(),
			TxIDRaw:          "INVALID",
		},
		ReqBody: `{"date": "2025-03-05", "merchant": "costco", "amount": 400.00, "detailed_category": 40}`,
	}
	t.Run(tc.Name, func(t *testing.T) {
		t.Parallel()
		env := testhelpers.SetupTestEnv(t)
		t.Cleanup(func() { env.Db.Close() })

		w := httptest.NewRecorder()

		req := httptest.NewRequest("POST", fmt.Sprintf("/transactions/%v", tc.TxIDRaw), bytes.NewBufferString(tc.ReqBody))
		req.Header.Set("Content-Type", "application/json")
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

		env.Router.POST("/transactions/:id", env.Middleware.RequestID(), env.Handlers.TxHandler.UpdateTransaction)
		env.Router.ServeHTTP(w, req)
		actualResponse := testhelpers.ProcessResponse(w, t)
		testhelpers.CheckHTTPResponse(t, w, tc.ExpectedError, tc.ExpectedStatusCode, tc.ExpectedResponse, actualResponse)
	})

}

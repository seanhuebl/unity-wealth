package transaction_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/models"
	"github.com/seanhuebl/unity-wealth/internal/testfixtures"
	"github.com/seanhuebl/unity-wealth/internal/testhelpers"
	"github.com/seanhuebl/unity-wealth/internal/testmodels"
)

func TestIntegrationGetTransactionsByUserID(t *testing.T) {
	t.Parallel()
	userID := uuid.New()
	txID := uuid.New()
	pagTxID := uuid.New()
	firstPageTests := []testmodels.GetAllTxByUserIDTestCase{
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
				Name:               "first page only, success",
				UserID:             userID,
				ExpectedStatusCode: http.StatusOK,
				ExpectedResponse: map[string]interface{}{
					"data": map[string]interface{}{
						"clamped":        false,
						"effective_size": int32(1),
						"has_more_data":  false,
						"next_cursor":    "",
						"transactions": []interface{}{
							map[string]interface{}{
								"id":                txID.String(),
								"user_id":           userID.String(),
								"date":              "2025-03-05T00:00:00Z",
								"merchant":          "costco",
								"amount":            125.98,
								"detailed_category": 40,
								"updated_at":        "0001-01-01T00:00:00Z",
							},
						},
					},
				},
			},
			PageSize: 1,
			MoreData: false,
		},
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
				Name:               "first page, more data, success",
				UserID:             userID,
				ExpectedStatusCode: http.StatusOK,
				ExpectedResponse: map[string]interface{}{
					"data": map[string]interface{}{
						"clamped":        false,
						"effective_size": int32(1),
						"has_more_data":  true,
						"next_cursor":    "token",
						"transactions": []interface{}{
							map[string]interface{}{
								"id":                txID.String(),
								"user_id":           userID.String(),
								"date":              "2025-03-05T00:00:00Z",
								"merchant":          "costco",
								"amount":            125.98,
								"detailed_category": 40,
								"updated_at":        "0001-01-01T00:00:00Z",
							},
						},
					},
				},
			},
			PageSize: 1,
			MoreData: true,
		},
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
				Name:               "no transactions: first page",
				UserID:             userID,
				ExpectedStatusCode: http.StatusOK,
				ExpectedResponse: map[string]interface{}{
					"data": map[string]interface{}{
						"clamped":        false,
						"effective_size": int32(1),
						"has_more_data":  false,
						"next_cursor":    "",
						"transactions":   []interface{}{},
					},
				},
			},
			PageSize: 1,
		},
	}

	for _, tc := range firstPageTests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			env := testhelpers.SetupTestEnv(t)
			defer env.Db.Close()
			if strings.Contains(tc.Name, "no transactions") {
				testhelpers.SeedTestUser(t, env.UserQ, userID, false)

			} else {
				testhelpers.SeedTestUser(t, env.UserQ, tc.UserID, false)
				testhelpers.IsTxFound(t, tc.BaseHTTPTestCase, txID, env)
				if tc.MoreData {
					testhelpers.SeedTestTransaction(t, env.TxQ, tc.UserID, pagTxID, &models.NewTxRequest{
						Date:             "2025-03-06",
						Merchant:         "costco",
						Amount:           125.98,
						DetailedCategory: 40,
					})
				}
			}
			env.Router.Use(func(c *gin.Context) {
				testhelpers.CheckForUserIDIssues(tc.Name, tc.UserID, c)
				c.Next()
			})
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", fmt.Sprintf("/transactions?limit=%d&cursor=%s", tc.PageSize, tc.Cursor), nil)
			env.Router.GET("/transactions", env.Middleware.RequestID(), env.Middleware.Paginate(), env.Handlers.TxHandler.GetTransactionsByUserID)

			env.Router.ServeHTTP(w, req)
			actualResponse := testhelpers.ProcessResponse(w, t)
			testhelpers.CheckHTTPResponse(t, w, tc.ExpectedError, tc.ExpectedStatusCode, tc.ExpectedResponse, actualResponse)
		})
	}
	pagTests := []testmodels.GetAllTxByUserIDTestCase{
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
				Name:               "paginated, only, success",
				UserID:             userID,
				ExpectedStatusCode: http.StatusOK,
				ExpectedResponse: map[string]interface{}{
					"data": map[string]interface{}{
						"clamped":        false,
						"effective_size": int32(1),
						"has_more_data":  false,
						"next_cursor":    "",
						"transactions": []interface{}{
							map[string]interface{}{
								"id":                pagTxID.String(),
								"user_id":           userID.String(),
								"date":              "2025-03-05T00:00:00Z",
								"merchant":          "costco",
								"amount":            125.98,
								"detailed_category": 40,
								"updated_at":        "0001-01-01T00:00:00Z",
							},
						},
					},
				},
			},
			NextCursor: "",
			PageSize:   1,
			MoreData:   false,
		},
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
				Name:               "paginated, more data, success",
				UserID:             userID,
				ExpectedStatusCode: http.StatusOK,
				ExpectedResponse: map[string]interface{}{
					"data": map[string]interface{}{
						"clamped":        false,
						"effective_size": int32(1),
						"has_more_data":  true,
						"next_cursor":    "token",
						"transactions": []interface{}{
							map[string]interface{}{
								"id":                pagTxID.String(),
								"user_id":           userID.String(),
								"date":              "2025-03-05T00:00:00Z",
								"merchant":          "costco",
								"amount":            125.98,
								"detailed_category": 40,
								"updated_at":        "0001-01-01T00:00:00Z",
							},
						},
					},
				},
			},
			NextCursor: "token",
			PageSize:   1,
			MoreData:   true,
		},
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
				Name:               "no transactions: paginated",
				UserID:             userID,
				ExpectedStatusCode: http.StatusOK,
				ExpectedResponse: map[string]interface{}{
					"data": map[string]interface{}{
						"clamped":        false,
						"effective_size": int32(1),
						"has_more_data":  false,
						"next_cursor":    "",
						"transactions":   []interface{}{},
					},
				},
			},
			PageSize: 1,
		},
	}
	for _, tc := range pagTests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			env := testhelpers.SetupTestEnv(t)
			defer env.Db.Close()
			if strings.Contains(tc.Name, "no transactions") {
				testhelpers.SeedTestUser(t, env.UserQ, userID, false)
			} else {
				testhelpers.SeedTestUser(t, env.UserQ, tc.UserID, false)
				testhelpers.IsTxFound(t, tc.BaseHTTPTestCase, pagTxID, env)
				if tc.MoreData {
					testhelpers.SeedTestTransaction(t, env.TxQ, tc.UserID, uuid.New(), &models.NewTxRequest{
						Date:             "2025-03-06",
						Merchant:         "costco",
						Amount:           125.98,
						DetailedCategory: 40,
					})
				}

			}
			env.Router.Use(func(c *gin.Context) {
				testhelpers.CheckForUserIDIssues(tc.Name, tc.UserID, c)
				c.Next()
			})
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", fmt.Sprintf("/transactions?limit=%d&cursor=%s", tc.PageSize, tc.Cursor), nil)
			env.Router.GET("/transactions", env.Middleware.RequestID(), env.Middleware.Paginate(), env.Handlers.TxHandler.GetTransactionsByUserID)

			env.Router.ServeHTTP(w, req)
			actualResponse := testhelpers.ProcessResponse(w, t)
			testhelpers.CheckHTTPResponse(t, w, tc.ExpectedError, tc.ExpectedStatusCode, tc.ExpectedResponse, actualResponse)
		})
	}

	errTests := []testmodels.GetAllTxByUserIDTestCase{
		{
			BaseHTTPTestCase: testfixtures.NilUserID(),
			PageSize:         1,
		},
		{
			BaseHTTPTestCase: testfixtures.InvalidUserID(),
			PageSize:         1,
		},
	}
	for _, tc := range errTests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			env := testhelpers.SetupTestEnv(t)
			t.Cleanup(func() { env.Db.Close() })

			env.Router.Use(func(c *gin.Context) {
				testhelpers.CheckForUserIDIssues(tc.Name, tc.UserID, c)
				c.Next()
			})
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", fmt.Sprintf("/transactions?limit=%d&cursor=%s", tc.PageSize, tc.Cursor), nil)
			env.Router.GET("/transactions", env.Middleware.RequestID(), env.Middleware.Paginate(), env.Handlers.TxHandler.GetTransactionsByUserID)

			env.Router.ServeHTTP(w, req)
			actualResponse := testhelpers.ProcessResponse(w, t)
			testhelpers.CheckHTTPResponse(t, w, tc.ExpectedError, tc.ExpectedStatusCode, tc.ExpectedResponse, actualResponse)
		})
	}
}

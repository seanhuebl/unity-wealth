package transaction_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	"github.com/seanhuebl/unity-wealth/internal/models"
	"github.com/seanhuebl/unity-wealth/internal/testfixtures"
	"github.com/seanhuebl/unity-wealth/internal/testhelpers"
	"github.com/seanhuebl/unity-wealth/internal/testmodels"
)

func TestIntegrationGetTransactionsByUserID(t *testing.T) {
	userID := uuid.New()
	txID := uuid.New()
	pagTxID := uuid.New()
	tests := []testmodels.GetAllTxByUserIDTestCase{
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
				Name:               "first page only, success",
				UserID:             userID,
				ExpectedStatusCode: http.StatusOK,
				ExpectedResponse: map[string]interface{}{
					"data": map[string]interface{}{
						"transactions": []interface{}{
							map[string]interface{}{
								"id":                txID.String(),
								"user_id":           userID.String(),
								"date":              "2025-03-05",
								"merchant":          "costco",
								"amount":            125.98,
								"detailed_category": 40,
							},
						},
						"next_cursor_date": "",
						"next_cursor_id":   "",
						"has_more_data":    false,
					},
				},
			},
			PageSize:      1,
			FirstPageTest: true,
			MoreData:      false,
		},
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
				Name:               "first page, more data, success",
				UserID:             userID,
				ExpectedStatusCode: http.StatusOK,
				ExpectedResponse: map[string]interface{}{
					"data": map[string]interface{}{
						"transactions": []interface{}{
							map[string]interface{}{
								"id":                txID.String(),
								"user_id":           userID.String(),
								"date":              "2025-03-05",
								"merchant":          "costco",
								"amount":            125.98,
								"detailed_category": 40,
							},
						},
						"next_cursor_date": "2025-03-05",
						"next_cursor_id":   txID.String(),
						"has_more_data":    true,
					},
				},
			},
			PageSize:      1,
			FirstPageTest: true,
			MoreData:      true,
		},
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
				Name:               "paginated, only, success",
				UserID:             userID,
				ExpectedStatusCode: http.StatusOK,
				ExpectedResponse: map[string]interface{}{
					"data": map[string]interface{}{
						"transactions": []interface{}{
							map[string]interface{}{
								"id":                pagTxID.String(),
								"user_id":           userID.String(),
								"date":              "2025-03-06",
								"merchant":          "costco",
								"amount":            125.98,
								"detailed_category": 40,
							},
						},
						"next_cursor_date": "",
						"next_cursor_id":   "",
						"has_more_data":    false,
					},
				},
			},
			CursorDate:    "2025-03-05",
			CursorID:      txID.String(),
			PageSize:      1,
			FirstPageTest: false,
			MoreData:      false,
		},
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
				Name:               "paginated, more data, success",
				UserID:             userID,
				ExpectedStatusCode: http.StatusOK,
				ExpectedResponse: map[string]interface{}{
					"data": map[string]interface{}{
						"transactions": []interface{}{
							map[string]interface{}{
								"id":                pagTxID.String(),
								"user_id":           userID.String(),
								"date":              "2025-03-06",
								"merchant":          "costco",
								"amount":            125.98,
								"detailed_category": 40,
							},
						},
						"next_cursor_date": "2025-03-06",
						"next_cursor_id":   pagTxID.String(),
						"has_more_data":    true,
					},
				},
			},
			CursorDate:    "2025-03-05",
			CursorID:      txID.String(),
			PageSize:      1,
			FirstPageTest: false,
			MoreData:      true,
		},
		{
			BaseHTTPTestCase: testfixtures.NilUserID,
		},
		{
			BaseHTTPTestCase: testfixtures.InvalidUserID,
		},
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
				Name:               "no transactions: first page",
				UserID:             userID,
				ExpectedStatusCode: http.StatusOK,
				ExpectedResponse: map[string]interface{}{
					"data": map[string]interface{}{
						"transactions":     []interface{}{},
						"next_cursor_date": "",
						"next_cursor_id":   "",
						"has_more_data":    false,
					},
				},
			},
			PageSize:      1,
			FirstPageTest: true,
		},
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
				Name:               "no transactions: paginated",
				UserID:             userID,
				ExpectedStatusCode: http.StatusOK,
				ExpectedResponse: map[string]interface{}{
					"data": map[string]interface{}{
						"transactions":     []interface{}{},
						"next_cursor_date": "",
						"next_cursor_id":   "",
						"has_more_data":    false,
					},
				},
			},
			PageSize:      1,
			FirstPageTest: false,
			CursorDate:    "2025-03-05",
			CursorID:      pagTxID.String(),
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			env := testhelpers.SetupTestEnv(t)
			defer env.Db.Close()
			if strings.Contains(tc.Name, "no transactions") && tc.FirstPageTest {
				testhelpers.SeedTestUser(t, env.UserQ, userID, false)

			} else if strings.Contains(tc.Name, "no transactions") && !tc.FirstPageTest {
				testhelpers.SeedTestUser(t, env.UserQ, userID, false)
				testhelpers.SeedTestCategories(t, env.Db)
				testhelpers.SeedTestTransaction(t, env.TxQ, userID, pagTxID, &models.NewTxRequest{
					Date:             "2025-03-05",
					Merchant:         "costco",
					Amount:           125.98,
					DetailedCategory: 40,
				})
			} else {
				if !strings.Contains(tc.Name, "unauthorized") {
					testhelpers.SeedTestUser(t, env.UserQ, tc.UserID, false)
					testhelpers.SeedTestCategories(t, env.Db)
					testhelpers.IsTxFound(t, tc.BaseHTTPTestCase, txID, env)
					if tc.MoreData || !tc.FirstPageTest {
						testhelpers.SeedTestTransaction(t, env.TxQ, tc.UserID, pagTxID, &models.NewTxRequest{
							Date:             "2025-03-06",
							Merchant:         "costco",
							Amount:           125.98,
							DetailedCategory: 40,
						})
					}
					if tc.MoreData && !tc.FirstPageTest {
						testhelpers.SeedTestTransaction(t, env.TxQ, tc.UserID, uuid.New(), &models.NewTxRequest{
							Date:             "2025-03-07",
							Merchant:         "costco",
							Amount:           125.98,
							DetailedCategory: 40,
						})
					}

				}
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/transactions", nil)
			env.Router.GET("/transactions", func(c *gin.Context) {
				c.Request = req
				c.Set(string(constants.CursorDateKey), tc.CursorDate)
				c.Set(string(constants.CursorIDKey), tc.CursorID)
				c.Set(string(constants.PageSizeKey), tc.PageSize)
				testhelpers.CheckForUserIDIssues(tc.Name, tc.UserID, c)
				env.Handlers.TxHandler.GetTransactionsByUserID(c)
			})
			env.Router.ServeHTTP(w, req)
			actualResponse := testhelpers.ProcessResponse(w, t)
			testhelpers.CheckHTTPResponse(t, w, tc.ExpectedError, tc.ExpectedStatusCode, tc.ExpectedResponse, actualResponse)
		})
	}
}

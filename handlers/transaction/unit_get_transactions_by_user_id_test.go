package transaction_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	htx "github.com/seanhuebl/unity-wealth/handlers/transaction"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	handlermocks "github.com/seanhuebl/unity-wealth/internal/mocks/handlers"
	"github.com/seanhuebl/unity-wealth/internal/models"
	"github.com/seanhuebl/unity-wealth/internal/services/transaction"
	"github.com/seanhuebl/unity-wealth/internal/testfixtures"
	"github.com/seanhuebl/unity-wealth/internal/testhelpers"
	"github.com/seanhuebl/unity-wealth/internal/testmodels"
	"github.com/stretchr/testify/mock"
)

func TestGetTransactionsByUserID(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	txID := uuid.New()
	date, _ := time.Parse("2006-01-02", "2025-03-19")
	successTests := []testmodels.GetAllTxByUserIDTestCase{
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
				Name:               "first page, more data, success",
				UserID:             userID,
				ExpectedStatusCode: http.StatusOK,
				ExpectedResponse: map[string]interface{}{
					"data": map[string]interface{}{
						"transactions": []interface{}{
							map[string]interface{}{
								"id":                txID,
								"user_id":           userID,
								"date":              date,
								"merchant":          "costco",
								"amount":            127.89,
								"detailed_category": 40,
							},
						},
						"next_cursor":    "token",
						"has_more_data":  true,
						"clamped":        false,
						"effective_size": 1,
					},
				},
			},
			NextCursor:    "token",
			PageSize:      1,
			FirstPageTest: true,
			MoreData:      true,
		},
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
				Name:               "first page only, success",
				UserID:             userID,
				ExpectedStatusCode: http.StatusOK,
				ExpectedResponse: map[string]interface{}{
					"data": map[string]interface{}{
						"transactions": []interface{}{
							map[string]interface{}{
								"id":                txID,
								"user_id":           userID,
								"date":              date,
								"merchant":          "costco",
								"amount":            127.89,
								"detailed_category": 40,
							},
						},
						"next_cursor":    "",
						"has_more_data":  false,
						"clamped":        false,
						"effective_size": 1,
					},
				},
			},
			PageSize:      1,
			FirstPageTest: true,
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
								"id":                txID,
								"user_id":           userID,
								"date":              date,
								"merchant":          "costco",
								"amount":            127.89,
								"detailed_category": 40,
							},
						},
						"next_cursor":    "token",
						"has_more_data":  true,
						"clamped":        false,
						"effective_size": 1,
					},
				},
			},
			NextCursor:    "token",
			PageSize:      1,
			FirstPageTest: false,
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
								"id":                txID,
								"user_id":           userID,
								"date":              date,
								"merchant":          "costco",
								"amount":            127.89,
								"detailed_category": 40,
							},
						},
						"next_cursor":    "",
						"has_more_data":  false,
						"clamped":        false,
						"effective_size": 1,
					},
				},
			},
			PageSize:      1,
			FirstPageTest: false,
			MoreData:      false,
		},
	}
	for _, tc := range successTests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			txs := []models.Tx{
				{
					ID:               txID,
					UserID:           tc.UserID,
					Date:             date,
					Merchant:         "costco",
					Amount:           127.89,
					DetailedCategory: 40,
				},
			}

			mockSvc := handlermocks.NewTransactionService(t)
			t.Cleanup(func() { mockSvc.AssertExpectations(t) })

			mockSvc.On(
				"ListUserTransactions",
				mock.Anything,
				tc.UserID,
				mock.AnythingOfType("string"),
				int64(tc.PageSize)).
				Return(transaction.ListTxResult{Transactions: txs}, nil).Once()

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/transactions", nil)

			h := htx.NewHandler(mockSvc)

			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set(string(constants.CursorKey), tc.NextCursor)

				c.Set(string(constants.LimitKey), tc.PageSize)
				testhelpers.CheckForUserIDIssues(tc.Name, tc.UserID, c)
				c.Next()
			})
			router.GET("/transactions", h.GetTransactionsByUserID)
			router.ServeHTTP(w, req)
			actualResponse := testhelpers.ProcessResponse(w, t)
			testhelpers.CheckHTTPResponse(t, w, tc.ExpectedError, tc.ExpectedStatusCode, tc.ExpectedResponse, actualResponse)
			mockSvc.AssertExpectations(t)
		})
	}

	errorTests := []testmodels.GetAllTxByUserIDTestCase{
		{
			BaseHTTPTestCase: testfixtures.NilUserID,
		},
		{
			BaseHTTPTestCase: testfixtures.InvalidUserID,
		},
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
				Name:               "error getting first page tx",
				UserID:             userID,
				ExpectedError:      "unable to get transactions",
				ExpectedStatusCode: http.StatusInternalServerError,
				ExpectedResponse: map[string]interface{}{
					"data": map[string]interface{}{
						"error": "unable to get transactions",
					},
				},
			},
			PageSize:        1,
			FirstPageTest:   true,
			GetFirstPageErr: errors.New("error getting transactions"),
		},
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
				Name:               "error getting paginated tx",
				UserID:             userID,
				ExpectedError:      "unable to get transactions",
				ExpectedStatusCode: http.StatusInternalServerError,
				ExpectedResponse: map[string]interface{}{
					"data": map[string]interface{}{
						"error": "unable to get transactions",
					},
				},
			},
			NextCursor:        "token",
			PageSize:          1,
			FirstPageTest:     false,
			GetTxPaginatedErr: errors.New("error getting transactions"),
		},
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
				Name:               "page size <= 0",
				UserID:             userID,
				ExpectedError:      "invalid page_size; must be > 0",
				ExpectedStatusCode: http.StatusBadRequest,
				ExpectedResponse: map[string]interface{}{
					"data": map[string]interface{}{
						"error": "invalid page_size; must be > 0",
					},
				},
			},
			PageSize: -1,
		},
	}
	for _, tc := range errorTests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			mockSvc := handlermocks.NewTransactionService(t)
			t.Cleanup(func() { mockSvc.AssertExpectations(t) })
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/transactions", nil)

			h := htx.NewHandler(mockSvc)

			if strings.Contains(tc.Name, "tx") {
				mockSvc.On(
					"ListUserTransactions",
					mock.Anything,
					tc.UserID,
					tc.NextCursor,
					int64(tc.PageSize)).
					Return(nil, "", "", false, errors.New(tc.ExpectedError))
			}

			router := gin.New()

			router.Use(func(c *gin.Context) {
				c.Set(string(constants.CursorKey), tc.NextCursor)
				c.Set(string(constants.LimitKey), tc.PageSize)
				testhelpers.CheckForUserIDIssues(tc.Name, tc.UserID, c)
				c.Next()
			})

			router.GET("/transactions", h.GetTransactionsByUserID)

			router.ServeHTTP(w, req)
			actualResponse := testhelpers.ProcessResponse(w, t)
			testhelpers.CheckHTTPResponse(t, w, tc.ExpectedError, tc.ExpectedStatusCode, tc.ExpectedResponse, actualResponse)
			mockSvc.AssertExpectations(t)
		})
	}
}

package transaction_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	htx "github.com/seanhuebl/unity-wealth/handlers/transaction"
	"github.com/seanhuebl/unity-wealth/internal/middleware"
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
	userID := uuid.New()
	txID := uuid.New()
	date, err := time.Parse(time.RFC3339, "2025-03-19T00:00:00Z")
	if err != nil {
		t.Fatalf("unable to parse: %v", err)
	}
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
								"id":                txID.String(),
								"user_id":           userID.String(),
								"date":              "2025-03-19T00:00:00Z",
								"merchant":          "costco",
								"amount":            127.89,
								"detailed_category": 40,
								"updated_at":        "0001-01-01T00:00:00Z",
							},
						},
						"next_cursor":    "token",
						"has_more_data":  true,
						"clamped":        false,
						"effective_size": int32(1),
					},
				},
			},
			NextCursor: "token",
			PageSize:   1,
			MoreData:   true,
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
								"id":                txID.String(),
								"user_id":           userID.String(),
								"date":              "2025-03-19T00:00:00Z",
								"merchant":          "costco",
								"amount":            127.89,
								"detailed_category": 40,
								"updated_at":        "0001-01-01T00:00:00Z",
							},
						},
						"next_cursor":    "",
						"has_more_data":  false,
						"clamped":        false,
						"effective_size": int32(1),
					},
				},
			},
			PageSize: 1,
			MoreData: false,
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
								"id":                txID.String(),
								"user_id":           userID.String(),
								"date":              "2025-03-19T00:00:00Z",
								"merchant":          "costco",
								"amount":            127.89,
								"detailed_category": 40,
								"updated_at":        "0001-01-01T00:00:00Z",
							},
						},
						"next_cursor":    "token",
						"has_more_data":  true,
						"clamped":        false,
						"effective_size": int32(1),
					},
				},
			},
			NextCursor: "token",
			PageSize:   1,
			MoreData:   true,
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
								"id":                txID.String(),
								"user_id":           userID.String(),
								"date":              "2025-03-19T00:00:00Z",
								"merchant":          "costco",
								"amount":            127.89,
								"detailed_category": 40,
								"updated_at":        "0001-01-01T00:00:00Z",
							},
						},
						"next_cursor":    "",
						"has_more_data":  false,
						"clamped":        false,
						"effective_size": int32(1),
					},
				},
			},
			PageSize: 1,
			MoreData: false,
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
				tc.PageSize).
				Return(transaction.ListTxResult{
					Transactions:  txs,
					NextCursor:    tc.Cursor,
					HasMoreData:   tc.MoreData,
					Clamped:       false,
					EffectiveSize: 1,
				}, nil).Once()

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", fmt.Sprintf("/transactions?limit=%d&cursor=%s", tc.PageSize, tc.Cursor), nil)

			h := htx.NewHandler(mockSvc)
			m := middleware.NewMiddleware(nil, nil)
			router := gin.New()
			router.Use(func(c *gin.Context) {
				testhelpers.CheckForUserIDIssues(tc.Name, tc.UserID, c)
				c.Next()
			})
			router.GET("/transactions", m.RequestID(), m.Paginate(), h.GetTransactionsByUserID)
			router.ServeHTTP(w, req)
			actualResponse := testhelpers.ProcessResponse(w, t)
			testhelpers.CheckHTTPResponse(t, w, tc.ExpectedError, tc.ExpectedStatusCode, tc.ExpectedResponse, actualResponse)
			mockSvc.AssertExpectations(t)
		})
	}

	errorTests := []testmodels.GetAllTxByUserIDTestCase{
		{
			BaseHTTPTestCase: testfixtures.NilUserID(),
			PageSize:         1,
		},
		{
			BaseHTTPTestCase: testfixtures.InvalidUserID(),
			PageSize:         1,
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
			GetTxPaginatedErr: errors.New("error getting transactions"),
		},
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
				Name:               "page size <= 0",
				UserID:             userID,
				ExpectedError:      "invalid limit; must be > 0",
				ExpectedStatusCode: http.StatusBadRequest,
				ExpectedResponse: map[string]interface{}{
					"data": map[string]interface{}{
						"error": "invalid limit; must be > 0",
					},
				},
			},
			PageSize:   -1,
			NextCursor: "",
		},
	}
	for _, tc := range errorTests {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			mockSvc := handlermocks.NewTransactionService(t)
			t.Cleanup(func() { mockSvc.AssertExpectations(t) })
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", fmt.Sprintf("/transactions?limit=%d&cursor=%s", tc.PageSize, tc.Cursor), nil)

			h := htx.NewHandler(mockSvc)

			if strings.Contains(tc.Name, "tx") {
				mockSvc.On(
					"ListUserTransactions",
					mock.Anything,
					tc.UserID,
					mock.AnythingOfType("string"),
					tc.PageSize).
					Return(transaction.ListTxResult{}, errors.New(tc.ExpectedError))
			}

			m := middleware.NewMiddleware(nil, nil)
			router := gin.New()
			router.Use(func(c *gin.Context) {
				testhelpers.CheckForUserIDIssues(tc.Name, tc.UserID, c)
				c.Next()
			})
			router.GET("/transactions", m.RequestID(), m.Paginate(), h.GetTransactionsByUserID)

			router.ServeHTTP(w, req)
			actualResponse := testhelpers.ProcessResponse(w, t)
			testhelpers.CheckHTTPResponse(t, w, tc.ExpectedError, tc.ExpectedStatusCode, tc.ExpectedResponse, actualResponse)
			mockSvc.AssertExpectations(t)
		})
	}
}

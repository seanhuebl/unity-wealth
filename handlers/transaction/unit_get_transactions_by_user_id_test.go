package transaction_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	htx "github.com/seanhuebl/unity-wealth/handlers/transaction"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	"github.com/seanhuebl/unity-wealth/internal/database"
	dbmocks "github.com/seanhuebl/unity-wealth/internal/mocks/database"
	"github.com/seanhuebl/unity-wealth/internal/services/transaction"
	"github.com/seanhuebl/unity-wealth/internal/testhelpers"
	"github.com/seanhuebl/unity-wealth/internal/testmodels"
	"github.com/stretchr/testify/mock"
)

func TestGetTransactionsByUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	txID := uuid.New()
	tests := []testmodels.GetAllTxByUserIDTestCase{
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
								"date":              "2025-03-19",
								"merchant":          "costco",
								"amount":            127.89,
								"detailed_category": 40,
							},
						},
						"next_cursor_date": "2025-03-19",
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
				Name:               "first page only, success",
				UserID:             userID,
				ExpectedStatusCode: http.StatusOK,
				ExpectedResponse: map[string]interface{}{
					"data": map[string]interface{}{
						"transactions": []interface{}{
							map[string]interface{}{
								"id":                txID.String(),
								"user_id":           userID.String(),
								"date":              "2025-03-19",
								"merchant":          "costco",
								"amount":            127.89,
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
				Name:               "paginated, more data, success",
				UserID:             userID,
				ExpectedStatusCode: http.StatusOK,
				ExpectedResponse: map[string]interface{}{
					"data": map[string]interface{}{
						"transactions": []interface{}{
							map[string]interface{}{
								"id":                txID.String(),
								"user_id":           userID.String(),
								"date":              "2025-03-19",
								"merchant":          "costco",
								"amount":            127.89,
								"detailed_category": 40,
							},
						},
						"next_cursor_date": "2025-03-19",
						"next_cursor_id":   txID.String(),
						"has_more_data":    true,
					},
				},
			},
			CursorDate:    "2025-03-19",
			CursorID:      txID.String(),
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
								"id":                txID.String(),
								"user_id":           userID.String(),
								"date":              "2025-03-19",
								"merchant":          "costco",
								"amount":            127.89,
								"detailed_category": 40,
							},
						},
						"next_cursor_date": "",
						"next_cursor_id":   "",
						"has_more_data":    false,
					},
				},
			},
			CursorDate:    "2025-03-19",
			CursorID:      txID.String(),
			PageSize:      1,
			FirstPageTest: false,
			MoreData:      false,
		},
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
				Name:               "unauthorized: user ID is uuid.NIL",
				UserID:             uuid.Nil,
				UserIDErr:          errors.New("user ID not found in context"),
				ExpectedError:      "unauthorized",
				ExpectedStatusCode: http.StatusUnauthorized,
			},
		},
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
				Name:               "unauthorized: user ID not UUID",
				UserID:             uuid.Nil,
				UserIDErr:          errors.New("user ID is not UUID"),
				ExpectedError:      "unauthorized",
				ExpectedStatusCode: http.StatusUnauthorized,
			},
		},
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
				Name:               "error getting first page tx",
				UserID:             userID,
				ExpectedError:      "unable to get transactions",
				ExpectedStatusCode: http.StatusInternalServerError,
				ExpectedResponse: map[string]interface{}{
					"error": "unable to get transactions",
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
					"error": "unable to get transactions",
				},
			},
			CursorDate:        "2025-03-19",
			CursorID:          txID.String(),
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
					"error": "invalid page_size; must be > 0",
				},
			},
			PageSize: -1,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			mockTxQ := dbmocks.NewTransactionQuerier(t)
			w := httptest.NewRecorder()
			svc := transaction.NewTransactionService(mockTxQ)
			req := httptest.NewRequest("GET", "/transactions", nil)
			firstPageTxSlice := []database.GetUserTransactionsFirstPageRow{
				{
					ID:                 txID.String(),
					UserID:             tc.BaseAccess().UserID.String(),
					TransactionDate:    "2025-03-19",
					Merchant:           "costco",
					AmountCents:        12789,
					DetailedCategoryID: 40,
				},
			}
			if tc.FirstPageTest && tc.MoreData {
				firstPageTxSlice = append(firstPageTxSlice, database.GetUserTransactionsFirstPageRow{
					ID:                 uuid.NewString(),
					UserID:             tc.BaseAccess().UserID.String(),
					TransactionDate:    "2025-03-20",
					Merchant:           "costco",
					AmountCents:        9999,
					DetailedCategoryID: 40,
				})
			}
			paginatedTxSlice := []database.GetUserTransactionsPaginatedRow{
				{
					ID:                 txID.String(),
					UserID:             tc.BaseAccess().UserID.String(),
					TransactionDate:    "2025-03-19",
					Merchant:           "costco",
					AmountCents:        12789,
					DetailedCategoryID: 40,
				},
			}
			if !tc.FirstPageTest && tc.MoreData {
				paginatedTxSlice = append(paginatedTxSlice, database.GetUserTransactionsPaginatedRow{
					ID:                 uuid.NewString(),
					UserID:             tc.BaseAccess().UserID.String(),
					TransactionDate:    "2025-03-20",
					Merchant:           "costco",
					AmountCents:        9999,
					DetailedCategoryID: 40,
				})
			}
			if tc.UserIDErr == nil && tc.PageSize > 0 {
				if tc.FirstPageTest {
					mockTxQ.On("GetUserTransactionsFirstPage", context.Background(), mock.AnythingOfType("database.GetUserTransactionsFirstPageParams")).
						Return(firstPageTxSlice, tc.GetFirstPageErr)
				} else {
					mockTxQ.On("GetUserTransactionsPaginated", context.Background(), mock.AnythingOfType("database.GetUserTransactionsPaginatedParams")).
						Return(paginatedTxSlice, tc.GetTxPaginatedErr)
				}
			}

			h := htx.NewHandler(svc)

			router := gin.New()
			router.GET("/transactions", func(c *gin.Context) {
				c.Request = req
				c.Set(string(constants.CursorDateKey), tc.CursorDate)
				c.Set(string(constants.CursorIDKey), tc.CursorID)
				c.Set(string(constants.PageSizeKey), tc.PageSize)
				if tc.Name == "unauthorized: user ID not UUID" {
					c.Set(string(constants.UserIDKey), "userID")
				} else {
					c.Set(string(constants.UserIDKey), tc.UserID)
				}
				h.GetTransactionsByUserID(c)
			})
			router.ServeHTTP(w, req)
			actualResponse := testhelpers.ProcessResponse(w, t)
			testhelpers.CheckTxHTTPResponse(t, w, tc, actualResponse)
			mockTxQ.AssertExpectations(t)
		})
	}
}

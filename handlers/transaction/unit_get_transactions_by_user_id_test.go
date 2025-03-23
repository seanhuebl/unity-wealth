package transaction

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	"github.com/seanhuebl/unity-wealth/internal/database"
	dbmocks "github.com/seanhuebl/unity-wealth/internal/mocks/database"
	"github.com/seanhuebl/unity-wealth/internal/services/transaction"
	"github.com/stretchr/testify/mock"
)

func TestGetTransactionsByUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	txID := uuid.New()
	tests := []GetAllTxByUserIDTestCase{
		{
			BaseHTTPTestCase: BaseHTTPTestCase{
				name:               "first page, more data, success",
				userID:             userID,
				expectedStatusCode: http.StatusOK,
				expectedResponse: map[string]interface{}{
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
			pageSize:      1,
			firstPageTest: true,
			moreData:      true,
		},
		{
			BaseHTTPTestCase: BaseHTTPTestCase{
				name:               "first page only, success",
				userID:             userID,
				expectedStatusCode: http.StatusOK,
				expectedResponse: map[string]interface{}{
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
			pageSize:      1,
			firstPageTest: true,
			moreData:      false,
		},
		{
			BaseHTTPTestCase: BaseHTTPTestCase{
				name:               "paginated, more data, success",
				userID:             userID,
				expectedStatusCode: http.StatusOK,
				expectedResponse: map[string]interface{}{
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
			cursorDate:    "2025-03-19",
			cursorID:      txID.String(),
			pageSize:      1,
			firstPageTest: false,
			moreData:      true,
		},
		{
			BaseHTTPTestCase: BaseHTTPTestCase{
				name:               "paginated, only, success",
				userID:             userID,
				expectedStatusCode: http.StatusOK,
				expectedResponse: map[string]interface{}{
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
			cursorDate:    "2025-03-19",
			cursorID:      txID.String(),
			pageSize:      1,
			firstPageTest: false,
			moreData:      false,
		},
		{
			BaseHTTPTestCase: BaseHTTPTestCase{
				name:               "unauthorized: user ID is uuid.NIL",
				userID:             uuid.Nil,
				userIDErr:          errors.New("user ID not found in context"),
				expectedError:      "unauthorized",
				expectedStatusCode: http.StatusUnauthorized,
			},
		},
		{
			BaseHTTPTestCase: BaseHTTPTestCase{
				name:               "unauthorized: user ID not UUID",
				userID:             uuid.Nil,
				userIDErr:          errors.New("user ID is not UUID"),
				expectedError:      "unauthorized",
				expectedStatusCode: http.StatusUnauthorized,
			},
		},
		{
			BaseHTTPTestCase: BaseHTTPTestCase{
				name:               "error getting first page tx",
				userID:             userID,
				expectedError:      "unable to get transactions",
				expectedStatusCode: http.StatusInternalServerError,
				expectedResponse: map[string]interface{}{
					"error": "unable to get transactions",
				},
			},
			pageSize:        1,
			firstPageTest:   true,
			getFirstPageErr: errors.New("error getting transactions"),
		},
		{
			BaseHTTPTestCase: BaseHTTPTestCase{
				name:               "error getting paginated tx",
				userID:             userID,
				expectedError:      "unable to get transactions",
				expectedStatusCode: http.StatusInternalServerError,
				expectedResponse: map[string]interface{}{
					"error": "unable to get transactions",
				},
			},
			cursorDate:        "2025-03-19",
			cursorID:          txID.String(),
			pageSize:          1,
			firstPageTest:     false,
			getTxPaginatedErr: errors.New("error getting transactions"),
		},
		{
			BaseHTTPTestCase: BaseHTTPTestCase{
				name:               "page size <= 0",
				userID:             userID,
				expectedError:      "invalid page_size; must be > 0",
				expectedStatusCode: http.StatusBadRequest,
				expectedResponse: map[string]interface{}{
					"error": "invalid page_size; must be > 0",
				},
			},
			pageSize: -1,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mockTxQ := dbmocks.NewTransactionQuerier(t)
			w := httptest.NewRecorder()
			svc := transaction.NewTransactionService(mockTxQ)
			req := httptest.NewRequest("GET", "/transactions", nil)
			firstPageTxSlice := []database.GetUserTransactionsFirstPageRow{
				{
					ID:                 txID.String(),
					UserID:             tc.BaseAccess().userID.String(),
					TransactionDate:    "2025-03-19",
					Merchant:           "costco",
					AmountCents:        12789,
					DetailedCategoryID: 40,
				},
			}
			if tc.firstPageTest && tc.moreData {
				firstPageTxSlice = append(firstPageTxSlice, database.GetUserTransactionsFirstPageRow{
					ID:                 uuid.NewString(),
					UserID:             tc.BaseAccess().userID.String(),
					TransactionDate:    "2025-03-20",
					Merchant:           "costco",
					AmountCents:        9999,
					DetailedCategoryID: 40,
				})
			}
			paginatedTxSlice := []database.GetUserTransactionsPaginatedRow{
				{
					ID:                 txID.String(),
					UserID:             tc.BaseAccess().userID.String(),
					TransactionDate:    "2025-03-19",
					Merchant:           "costco",
					AmountCents:        12789,
					DetailedCategoryID: 40,
				},
			}
			if !tc.firstPageTest && tc.moreData {
				paginatedTxSlice = append(paginatedTxSlice, database.GetUserTransactionsPaginatedRow{
					ID:                 uuid.NewString(),
					UserID:             tc.BaseAccess().userID.String(),
					TransactionDate:    "2025-03-20",
					Merchant:           "costco",
					AmountCents:        9999,
					DetailedCategoryID: 40,
				})
			}
			if tc.userIDErr == nil && tc.pageSize > 0 {
				if tc.firstPageTest {
					mockTxQ.On("GetUserTransactionsFirstPage", context.Background(), mock.AnythingOfType("database.GetUserTransactionsFirstPageParams")).
						Return(firstPageTxSlice, tc.getFirstPageErr)
				} else {
					mockTxQ.On("GetUserTransactionsPaginated", context.Background(), mock.AnythingOfType("database.GetUserTransactionsPaginatedParams")).
						Return(paginatedTxSlice, tc.getTxPaginatedErr)
				}
			}

			h := NewHandler(svc)

			router := gin.New()
			router.GET("/transactions", func(c *gin.Context) {
				c.Request = req
				c.Set(string(constants.CursorDateKey), tc.cursorDate)
				c.Set(string(constants.CursorIDKey), tc.cursorID)
				c.Set(string(constants.PageSizeKey), tc.pageSize)
				if tc.name == "unauthorized: user ID not UUID" {
					c.Set(string(constants.UserIDKey), "userID")
				} else {
					c.Set(string(constants.UserIDKey), tc.userID)
				}
				h.GetTransactionsByUserID(c)
			})
			router.ServeHTTP(w, req)
			actualResponse := processResponse(w, t)
			checkTxHTTPResponse(t, w, tc, actualResponse)
			mockTxQ.AssertExpectations(t)
		})
	}
}

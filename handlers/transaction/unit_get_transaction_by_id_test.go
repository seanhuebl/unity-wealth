package transaction

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	"github.com/seanhuebl/unity-wealth/internal/database"
	dbmocks "github.com/seanhuebl/unity-wealth/internal/mocks/database"
	"github.com/seanhuebl/unity-wealth/internal/services/transaction"
)

func TestGetTxByID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []GetTxTestCase{
		{
			BaseHTTPTestCase: BaseHTTPTestCase{

				name:               "success",
				userID:             uuid.New(),
				expectedStatusCode: http.StatusOK,
				expectedResponse: map[string]interface{}{
					"data": map[string]interface{}{
						"date":              "2025-03-05",
						"merchant":          "costco",
						"amount":            125.98,
						"detailed_category": 40,
					},
				},
			},
			txID: uuid.NewString(),
		},
		{
			BaseHTTPTestCase: BaseHTTPTestCase{
				name:               "unauthorized: user ID is uuid.NIL",
				userID:             uuid.Nil,
				userIDErr:          errors.New("user ID not found in context"),
				expectedError:      "unauthorized",
				expectedStatusCode: http.StatusUnauthorized,
			},
			txID: uuid.NewString(),
		},
		{
			BaseHTTPTestCase: BaseHTTPTestCase{
				name:               "unauthorized: user ID not UUID",
				userID:             uuid.Nil,
				userIDErr:          errors.New("user ID is not UUID"),
				expectedError:      "unauthorized",
				expectedStatusCode: http.StatusUnauthorized,
			},
			txID: uuid.NewString(),
		},
		{
			BaseHTTPTestCase: BaseHTTPTestCase{

				name:               "error getting tx",
				userID:             uuid.New(),
				expectedError:      "unable to get transaction",
				expectedStatusCode: http.StatusInternalServerError,
			},
			txID:  uuid.NewString(),
			txErr: errors.New("error getting transaction"),
		},
		{
			BaseHTTPTestCase: BaseHTTPTestCase{
				name:               "invalid txID in req",
				userID:             uuid.New(),
				expectedError:      "invalid id",
				expectedStatusCode: http.StatusBadRequest,
			},
			txID: "",
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mockTxQ := dbmocks.NewTransactionQuerier(t)
			w := httptest.NewRecorder()
			svc := transaction.NewTransactionService(mockTxQ)
			req := httptest.NewRequest("GET", fmt.Sprintf("/transactions/%v", tc.txID), nil)

			dummyRow := database.GetUserTransactionByIDRow{
				ID:                 tc.txID,
				UserID:             tc.userID.String(),
				TransactionDate:    "2025-03-05",
				Merchant:           "costco",
				AmountCents:        12598,
				DetailedCategoryID: 40,
			}

			if tc.userIDErr == nil && tc.txID != "" {
				mockTxQ.On("GetUserTransactionByID", context.Background(), database.GetUserTransactionByIDParams{
					UserID: tc.userID.String(),
					ID:     tc.txID,
				}).Return(dummyRow, tc.txErr)
			}

			h := NewHandler(svc)

			router := gin.New()
			if tc.txID == "" {
				c, _ := gin.CreateTestContext(w)
				c.Request = req
				c.Params = gin.Params{{Key: "id", Value: ""}}
				if tc.name == "unauthorized: user ID not UUID" {
					c.Set(string(constants.UserIDKey), "userID")
				} else {
					c.Set(string(constants.UserIDKey), tc.userID)
				}
				h.GetTransactionByID(c)
			} else {

				router.GET("/transactions/:id", func(c *gin.Context) {
					if tc.name == "unauthorized: user ID not UUID" {
						c.Set(string(constants.UserIDKey), "userID")
					} else {
						c.Set(string(constants.UserIDKey), tc.userID)
					}
					h.GetTransactionByID(c)
				})
				router.ServeHTTP(w, req)
			}

			actualResponse := processResponse(w, t)
			checkTxHTTPResponse(t, w, tc, actualResponse)
			mockTxQ.AssertExpectations(t)
		})
	}
}

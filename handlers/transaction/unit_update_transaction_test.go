package transaction

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/seanhuebl/unity-wealth/internal/helpers"
	dbmocks "github.com/seanhuebl/unity-wealth/internal/mocks/database"
	"github.com/seanhuebl/unity-wealth/internal/services/transaction"
	"github.com/stretchr/testify/mock"
)

func TestUpdateTransaction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []UpdateTxTestCase{
		{
			GetTxTestCase: GetTxTestCase{
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
			reqBody: `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
		},
		{
			GetTxTestCase: GetTxTestCase{
				BaseHTTPTestCase: BaseHTTPTestCase{

					name:               "unauthorized: user ID is uuid.NIL",
					userID:             uuid.Nil,
					userIDErr:          errors.New("user ID not found in context"),
					expectedError:      "unauthorized",
					expectedStatusCode: http.StatusUnauthorized,
				},
				txID: uuid.NewString(),
			},
			reqBody: `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
		},
		{
			GetTxTestCase: GetTxTestCase{
				BaseHTTPTestCase: BaseHTTPTestCase{
					name:               "unauthorized: user ID not UUID",
					userID:             uuid.Nil,
					userIDErr:          errors.New("user ID is not UUID"),
					expectedError:      "unauthorized",
					expectedStatusCode: http.StatusUnauthorized,
				},
				txID: uuid.NewString(),
			},
			reqBody: `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
		},
		{
			GetTxTestCase: GetTxTestCase{
				BaseHTTPTestCase: BaseHTTPTestCase{
					name:               "invalid request body",
					userID:             uuid.New(),
					expectedError:      "invalid request body",
					expectedStatusCode: http.StatusBadRequest,
				},
				txID: uuid.NewString(),
			},
			reqBody: `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40`,
		},
		{
			GetTxTestCase: GetTxTestCase{
				BaseHTTPTestCase: BaseHTTPTestCase{
					name:               "failed to update tx",
					userID:             uuid.New(),
					expectedError:      "failed to update transaction",
					expectedStatusCode: http.StatusInternalServerError,
				},
				txID: uuid.NewString(),
			},
			reqBody:     `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
			updateTxErr: errors.New("update err"),
		},
		{
			GetTxTestCase: GetTxTestCase{
				BaseHTTPTestCase: BaseHTTPTestCase{
					name:               "invalid txID in req",
					userID:             uuid.New(),
					expectedError:      "invalid id",
					expectedStatusCode: http.StatusBadRequest,
				},
				txID: "",
			},
			reqBody: `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockTxQ := dbmocks.NewTransactionQuerier(t)
			w := httptest.NewRecorder()
			svc := transaction.NewTransactionService(mockTxQ)
			req := httptest.NewRequest("POST", fmt.Sprintf("/transactions/%v", tc.txID), bytes.NewBufferString(tc.reqBody))
			req.Header.Set("Content-Type", "application/json")

			if tc.name == "unauthorized:user ID not UUID" {
				req = req.WithContext(context.WithValue(req.Context(), constants.UserIDKey, "userID"))
			} else {
				req = req.WithContext(context.WithValue(req.Context(), constants.UserIDKey, tc.userID))
			}

			dummyRow := database.UpdateTransactionByIDRow{
				ID:                 tc.txID,
				TransactionDate:    "2025-03-05",
				Merchant:           "costco",
				AmountCents:        helpers.ConvertToCents(125.98),
				DetailedCategoryID: 40,
			}

			if json.Valid([]byte(tc.reqBody)) && tc.userIDErr == nil && tc.txID != "" {
				mockTxQ.On("UpdateTransactionByID", req.Context(), mock.AnythingOfType("database.UpdateTransactionByIDParams")).
					Return(dummyRow, tc.updateTxErr)
			}
			h := NewHandler(svc)

			router := gin.New()
			if tc.txID == "" {
				c, _ := gin.CreateTestContext(w)
				c.Request = req
				c.Params = gin.Params{{Key: "id", Value: ""}}
				h.UpdateTransaction(c)
			} else {
				router.POST("/transactions/:id", h.UpdateTransaction)
				router.ServeHTTP(w, req)

			}

			actualResponse := processResponse(w, t)

			checkTxHTTPResponse(t, w, tc, actualResponse)
			mockTxQ.AssertExpectations(t)
		})
	}
}

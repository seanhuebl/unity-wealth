package transaction

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	dbmocks "github.com/seanhuebl/unity-wealth/internal/mocks/database"
	"github.com/seanhuebl/unity-wealth/internal/services/transaction"
	"github.com/stretchr/testify/mock"
)

func TestNewTx(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []CreateTxTestCase{
		{
			BaseHTTPTestCase: BaseHTTPTestCase{
				name:               "success",
				userID:             uuid.New(),
				expectedStatusCode: http.StatusCreated,
				expectedResponse: map[string]interface{}{
					"data": map[string]interface{}{
						"date":              "2025-03-05",
						"merchant":          "costco",
						"amount":            125.98,
						"detailed_category": 40,
					},
				},
			},
			reqBody: `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
		},
		{
			BaseHTTPTestCase: BaseHTTPTestCase{
				name:               "unauthorized: user ID is uuid.NIL",
				userID:             uuid.Nil,
				userIDErr:          errors.New("user ID not found in context"),
				expectedError:      "unauthorized",
				expectedStatusCode: http.StatusUnauthorized,
			},
			reqBody: `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
		},
		{
			BaseHTTPTestCase: BaseHTTPTestCase{
				name:               "unauthorized: user ID not UUID",
				userID:             uuid.Nil,
				userIDErr:          errors.New("user ID is not UUID"),
				expectedError:      "unauthorized",
				expectedStatusCode: http.StatusUnauthorized,
			},
			reqBody: `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
		},
		{
			BaseHTTPTestCase: BaseHTTPTestCase{
				name:               "invalid request body",
				userID:             uuid.New(),
				expectedError:      "invalid request body",
				expectedStatusCode: http.StatusBadRequest,
			},
			reqBody: `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40`,
		},
		{
			BaseHTTPTestCase: BaseHTTPTestCase{
				name:               "failed to create transaction",
				userID:             uuid.New(),
				expectedError:      "failed to create transaction",
				expectedStatusCode: http.StatusInternalServerError,
			},
			reqBody:     `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
			createTxErr: errors.New("create tx error"),
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mockTxQ := dbmocks.NewTransactionQuerier(t)

			w := httptest.NewRecorder()

			svc := transaction.NewTransactionService(mockTxQ)

			req := httptest.NewRequest("POST", "/transactions", bytes.NewBufferString(tc.reqBody))
			req.Header.Set("Content-Type", "application/json")

			if tc.name == "unauthorized: user ID not UUID" {
				req = req.WithContext(context.WithValue(req.Context(), constants.UserIDKey, "userID"))
			} else {
				req = req.WithContext(context.WithValue(req.Context(), constants.UserIDKey, tc.userID))
			}

			if json.Valid([]byte(tc.reqBody)) && tc.userIDErr == nil {
				mockTxQ.On("CreateTransaction", req.Context(), mock.AnythingOfType("database.CreateTransactionParams")).Return(tc.createTxErr)
			}
			h := NewHandler(svc)

			router := gin.New()
			router.POST("/transactions", h.NewTransaction)
			router.ServeHTTP(w, req)

			actualResponse := processResponse(w, t)

			checkTxHTTPResponse(t, w, tc, actualResponse)
			mockTxQ.AssertExpectations(t)
		})
	}
}

package transaction_test

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
	htx "github.com/seanhuebl/unity-wealth/handlers/transaction"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	dbmocks "github.com/seanhuebl/unity-wealth/internal/mocks/database"
	"github.com/seanhuebl/unity-wealth/internal/services/transaction"
	"github.com/seanhuebl/unity-wealth/internal/testhelpers"
	"github.com/seanhuebl/unity-wealth/internal/testmodels"
	"github.com/stretchr/testify/mock"
)

func TestNewTx(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []testmodels.CreateTxTestCase{
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
				Name:               "success",
				UserID:             uuid.New(),
				ExpectedStatusCode: http.StatusCreated,
				ExpectedResponse: map[string]interface{}{
					"data": map[string]interface{}{
						"date":              "2025-03-05",
						"merchant":          "costco",
						"amount":            125.98,
						"detailed_category": 40,
					},
				},
			},
			ReqBody: `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
		},
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
				Name:               "unauthorized: user ID is uuid.NIL",
				UserID:             uuid.Nil,
				UserIDErr:          errors.New("user ID not found in context"),
				ExpectedError:      "unauthorized",
				ExpectedStatusCode: http.StatusUnauthorized,
			},
			ReqBody: `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
		},
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
				Name:               "unauthorized: user ID not UUID",
				UserID:             uuid.Nil,
				UserIDErr:          errors.New("user ID is not UUID"),
				ExpectedError:      "unauthorized",
				ExpectedStatusCode: http.StatusUnauthorized,
			},
			ReqBody: `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
		},
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
				Name:               "invalid request body",
				UserID:             uuid.New(),
				ExpectedError:      "invalid request body",
				ExpectedStatusCode: http.StatusBadRequest,
			},
			ReqBody: `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40`,
		},
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
				Name:               "failed to create transaction",
				UserID:             uuid.New(),
				ExpectedError:      "failed to create transaction",
				ExpectedStatusCode: http.StatusInternalServerError,
			},
			ReqBody:     `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
			CreateTxErr: errors.New("create tx error"),
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			mockTxQ := dbmocks.NewTransactionQuerier(t)

			w := httptest.NewRecorder()

			svc := transaction.NewTransactionService(mockTxQ)

			req := httptest.NewRequest("POST", "/transactions", bytes.NewBufferString(tc.ReqBody))
			req.Header.Set("Content-Type", "application/json")

			if tc.Name == "unauthorized: user ID not UUID" {
				req = req.WithContext(context.WithValue(req.Context(), constants.UserIDKey, "userID"))
			} else {
				req = req.WithContext(context.WithValue(req.Context(), constants.UserIDKey, tc.UserID))
			}

			if json.Valid([]byte(tc.ReqBody)) && tc.UserIDErr == nil {
				mockTxQ.On("CreateTransaction", req.Context(), mock.AnythingOfType("database.CreateTransactionParams")).Return(tc.CreateTxErr)
			}
			h := htx.NewHandler(svc)

			router := gin.New()
			router.POST("/transactions", h.NewTransaction)
			router.ServeHTTP(w, req)

			actualResponse := testhelpers.ProcessResponse(w, t)

			testhelpers.CheckTxHTTPResponse(t, w, tc, actualResponse)
			mockTxQ.AssertExpectations(t)
		})
	}
}

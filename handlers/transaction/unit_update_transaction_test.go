package transaction_test

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
	htx "github.com/seanhuebl/unity-wealth/handlers/transaction"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/seanhuebl/unity-wealth/internal/helpers"
	dbmocks "github.com/seanhuebl/unity-wealth/internal/mocks/database"
	"github.com/seanhuebl/unity-wealth/internal/services/transaction"
	"github.com/seanhuebl/unity-wealth/internal/testhelpers"
	"github.com/seanhuebl/unity-wealth/internal/testmodels"
	"github.com/stretchr/testify/mock"
)

func TestUpdateTransaction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []testmodels.UpdateTxTestCase{
		{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
					Name:               "success",
					UserID:             uuid.New(),
					ExpectedStatusCode: http.StatusOK,
					ExpectedResponse: map[string]interface{}{
						"data": map[string]interface{}{
							"date":              "2025-03-05",
							"merchant":          "costco",
							"amount":            125.98,
							"detailed_category": 40,
						},
					},
				},
				TxID: uuid.NewString(),
			},
			ReqBody: `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
		},
		{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testmodels.BaseHTTPTestCase{

					Name:               "unauthorized: user ID is uuid.NIL",
					UserID:             uuid.Nil,
					UserIDErr:          errors.New("user ID not found in context"),
					ExpectedError:      "unauthorized",
					ExpectedStatusCode: http.StatusUnauthorized,
				},
				TxID: uuid.NewString(),
			},
			ReqBody: `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
		},
		{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
					Name:               "unauthorized: user ID not UUID",
					UserID:             uuid.Nil,
					UserIDErr:          errors.New("user ID is not UUID"),
					ExpectedError:      "unauthorized",
					ExpectedStatusCode: http.StatusUnauthorized,
				},
				TxID: uuid.NewString(),
			},
			ReqBody: `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
		},
		{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
					Name:               "invalid request body",
					UserID:             uuid.New(),
					ExpectedError:      "invalid request body",
					ExpectedStatusCode: http.StatusBadRequest,
				},
				TxID: uuid.NewString(),
			},
			ReqBody: `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40`,
		},
		{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
					Name:               "failed to update tx",
					UserID:             uuid.New(),
					ExpectedError:      "failed to update transaction",
					ExpectedStatusCode: http.StatusInternalServerError,
				},
				TxID: uuid.NewString(),
			},
			ReqBody:     `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
			UpdateTxErr: errors.New("update err"),
		},
		{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
					Name:               "invalid txID in req",
					UserID:             uuid.New(),
					ExpectedError:      "invalid id",
					ExpectedStatusCode: http.StatusBadRequest,
				},
				TxID: "",
			},
			ReqBody: `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			mockTxQ := dbmocks.NewTransactionQuerier(t)
			w := httptest.NewRecorder()
			svc := transaction.NewTransactionService(mockTxQ)
			req := httptest.NewRequest("POST", fmt.Sprintf("/transactions/%v", tc.TxID), bytes.NewBufferString(tc.ReqBody))
			req.Header.Set("Content-Type", "application/json")

			if tc.Name == "unauthorized:user ID not UUID" {
				req = req.WithContext(context.WithValue(req.Context(), constants.UserIDKey, "userID"))
			} else {
				req = req.WithContext(context.WithValue(req.Context(), constants.UserIDKey, tc.UserID))
			}

			dummyRow := database.UpdateTransactionByIDRow{
				ID:                 tc.TxID,
				TransactionDate:    "2025-03-05",
				Merchant:           "costco",
				AmountCents:        helpers.ConvertToCents(125.98),
				DetailedCategoryID: 40,
			}

			if json.Valid([]byte(tc.ReqBody)) && tc.UserIDErr == nil && tc.TxID != "" {
				mockTxQ.On("UpdateTransactionByID", req.Context(), mock.AnythingOfType("database.UpdateTransactionByIDParams")).
					Return(dummyRow, tc.UpdateTxErr)
			}
			h := htx.NewHandler(svc)

			router := gin.New()
			if tc.TxID == "" {
				c, _ := gin.CreateTestContext(w)
				c.Request = req
				c.Params = gin.Params{{Key: "id", Value: ""}}
				h.UpdateTransaction(c)
			} else {
				router.POST("/transactions/:id", h.UpdateTransaction)
				router.ServeHTTP(w, req)

			}

			actualResponse := testhelpers.ProcessResponse(w, t)

			testhelpers.CheckTxHTTPResponse(t, w, tc, actualResponse)
			mockTxQ.AssertExpectations(t)
		})
	}
}

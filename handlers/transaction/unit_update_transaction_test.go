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
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/seanhuebl/unity-wealth/internal/helpers"
	dbmocks "github.com/seanhuebl/unity-wealth/internal/mocks/database"
	"github.com/seanhuebl/unity-wealth/internal/services/transaction"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdateTransaction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name               string
		userID             uuid.UUID
		userIDErr          error
		reqBody            string
		txID               string
		updateTxErr        error
		expectedErr        string
		expectedStatusCode int
		expectedResponse   map[string]interface{}
	}{
		{
			name:               "success",
			userID:             uuid.New(),
			txID:               uuid.NewString(),
			reqBody:            `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
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
		{
			name:               "unauthorized: user ID is uuid.NIL",
			userID:             uuid.Nil,
			userIDErr:          errors.New("user ID not found in context"),
			reqBody:            `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
			txID:               uuid.NewString(),
			expectedErr:        "unauthorized",
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "unauthorized: user ID not UUID",
			userID:             uuid.Nil,
			userIDErr:          errors.New("user ID is not UUID"),
			reqBody:            `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
			txID:               uuid.NewString(),
			expectedErr:        "unauthorized",
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "invalid request body",
			userID:             uuid.New(),
			reqBody:            `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40`,
			txID:               uuid.NewString(),
			expectedErr:        "invalid request body",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "failed to update tx",
			userID:             uuid.New(),
			reqBody:            `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
			txID:               uuid.NewString(),
			updateTxErr:        errors.New("update err"),
			expectedErr:        "failed to update transaction",
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:               "invalid txID in req",
			userID:             uuid.New(),
			reqBody:            `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
			txID:               "",
			expectedErr:        "invalid id",
			expectedStatusCode: http.StatusBadRequest,
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

			var actualResponse map[string]interface{}

			err := json.Unmarshal(w.Body.Bytes(), &actualResponse)
			require.NoError(t, err)

			actualResponse = convertResponseFloatToInt(actualResponse)

			if tc.expectedErr != "" {
				require.Contains(t, actualResponse["error"].(string), tc.expectedErr)
			} else {
				if diff := cmp.Diff(tc.expectedResponse, actualResponse); diff != "" {
					t.Errorf("response mismatch (-want +got)\n%s", diff)
				}
			}
			if diff := cmp.Diff(tc.expectedStatusCode, w.Code); diff != "" {
				t.Errorf("status code mismatch (-want +got)\n%s", diff)
			}
			mockTxQ.AssertExpectations(t)
		})
	}
}

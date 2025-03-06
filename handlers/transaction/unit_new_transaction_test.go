package transaction

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	dbmocks "github.com/seanhuebl/unity-wealth/internal/mocks/database"
	"github.com/seanhuebl/unity-wealth/internal/services/transaction"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewTx(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name               string
		userID             uuid.UUID
		userIDErr          error
		reqBody            string
		createTxErr        error
		expectedError      string
		expectedStatusCode int
		expectedResponse   map[string]interface{}
	}{
		{
			name:               "success",
			userID:             uuid.New(),
			reqBody:            `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
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
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mockTxQ := dbmocks.NewTransactionQuerier(t)

			w := httptest.NewRecorder()

			svc := transaction.NewTransactionService(mockTxQ)

			req := httptest.NewRequest("POST", "/transactions", bytes.NewBufferString(tc.reqBody))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(context.WithValue(req.Context(), constants.UserIDKey, tc.userID))

			if json.Valid([]byte(tc.reqBody)) {
				mockTxQ.On("CreateTransaction", req.Context(), mock.AnythingOfType("database.CreateTransactionParams")).Return(tc.createTxErr)
			}
			h := NewHandler(svc)

			router := gin.New()
			router.POST("/transactions", h.NewTransaction)
			router.ServeHTTP(w, req)

			var actualResponse map[string]interface{}

			err := json.Unmarshal(w.Body.Bytes(), &actualResponse)
			require.NoError(t, err)

			if tc.expectedError != "" {
				require.Contains(t, actualResponse["error"].(string), tc.expectedError)
			} else {
				if diff := cmp.Diff(tc.expectedResponse, actualResponse); diff != "" {
					t.Errorf("response mismatch (-want, +got)\n%s", diff)
				}
			}
			if diff := cmp.Diff(tc.expectedStatusCode, w.Code); diff != "" {
				t.Errorf("status code mismatch (-want, +got)\n%s", diff)
			}
			mockTxQ.AssertExpectations(t)
		})
	}
}

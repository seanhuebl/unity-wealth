package transaction

import (
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
	dbmocks "github.com/seanhuebl/unity-wealth/internal/mocks/database"
	"github.com/seanhuebl/unity-wealth/internal/services/transaction"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDeleteTransaction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name               string
		userID             uuid.UUID
		userIDErr          error
		txID               string
		txErr              error
		expectedErr        string
		expectedStatusCode int
		expectedResponse   map[string]interface{}
	}{
		{
			name:               "success",
			userID:             uuid.New(),
			txID:               uuid.NewString(),
			expectedStatusCode: http.StatusOK,
			expectedResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"transaction_deleted": "success",
				},
			},
		},
		{
			name:               "unauthorized: user ID is uuid.NIL",
			userID:             uuid.Nil,
			txID:               uuid.NewString(),
			userIDErr:          errors.New("user ID not found in context"),
			expectedErr:        "unauthorized",
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "unauthorized: user ID not UUID",
			userID:             uuid.Nil,
			userIDErr:          errors.New("user ID is not UUID"),
			txID:               uuid.NewString(),
			expectedErr:        "unauthorized",
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "error deleting tx",
			userID:             uuid.New(),
			txID:               uuid.NewString(),
			txErr:              errors.New("error deleting transaction"),
			expectedErr:        "error deleting transaction",
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:               "invalid txID in req",
			userID:             uuid.New(),
			txID:               "",
			expectedErr:        "invalid id",
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mockTxQ := dbmocks.NewTransactionQuerier(t)
			w := httptest.NewRecorder()
			svc := transaction.NewTransactionService(mockTxQ)
			req := httptest.NewRequest("DELETE", fmt.Sprintf("/transactions/%v", tc.txID), nil)

			if tc.userIDErr == nil && tc.txID != "" {
				mockTxQ.On("DeleteTransactionByID", context.Background(), mock.AnythingOfType("database.DeleteTransactionByIDParams")).Return(tc.txErr)
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
				h.DeleteTransaction(c)
			} else {
				router.DELETE("/transactions/:id", func(c *gin.Context) {
					if tc.name == "unauthorized: user ID not UUID" {
						c.Set(string(constants.UserIDKey), "userID")
					} else {
						c.Set(string(constants.UserIDKey), tc.userID)
					}
					h.DeleteTransaction(c)
				})
				router.ServeHTTP(w, req)
			}

			var actualResponse map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &actualResponse)
			require.NoError(t, err)

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

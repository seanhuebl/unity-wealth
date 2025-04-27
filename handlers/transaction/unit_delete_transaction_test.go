package transaction_test

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
	htx "github.com/seanhuebl/unity-wealth/handlers/transaction"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	dbmocks "github.com/seanhuebl/unity-wealth/internal/mocks/database"
	"github.com/seanhuebl/unity-wealth/internal/services/transaction"
	"github.com/seanhuebl/unity-wealth/internal/testmodels"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDeleteTransaction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []testmodels.DeleteTxTestCase{
		{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
					Name:               "success",
					UserID:             uuid.New(),
					ExpectedStatusCode: http.StatusOK,
					ExpectedResponse: map[string]interface{}{
						"data": map[string]interface{}{
							"transaction_deleted": "success",
						},
					},
				},
				TxID: uuid.NewString(),
			},
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
		},
		{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testmodels.BaseHTTPTestCase{

					Name:               "error deleting tx",
					UserID:             uuid.New(),
					ExpectedError:      "error deleting transaction",
					ExpectedStatusCode: http.StatusInternalServerError,
				},
				TxID:  uuid.NewString(),
				TxErr: errors.New("error deleting transaction"),
			},
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
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			mockTxQ := dbmocks.NewTransactionQuerier(t)
			w := httptest.NewRecorder()
			svc := transaction.NewTransactionService(mockTxQ)
			req := httptest.NewRequest("DELETE", fmt.Sprintf("/transactions/%v", tc.TxID), nil)

			if tc.UserIDErr == nil && tc.TxID != "" {
				mockTxQ.On("DeleteTransactionByID", context.Background(), mock.AnythingOfType("database.DeleteTransactionByIDParams")).Return(tc.TxErr)
			}

			h := htx.NewHandler(svc)

			router := gin.New()
			if tc.TxID == "" {
				c, _ := gin.CreateTestContext(w)
				c.Request = req
				c.Params = gin.Params{{Key: "id", Value: ""}}
				if tc.Name == "unauthorized: user ID not UUID" {
					c.Set(string(constants.UserIDKey), "userID")
				} else {
					c.Set(string(constants.UserIDKey), tc.UserID)
				}
				h.DeleteTransaction(c)
			} else {
				router.DELETE("/transactions/:id", func(c *gin.Context) {
					if tc.Name == "unauthorized: user ID not UUID" {
						c.Set(string(constants.UserIDKey), "userID")
					} else {
						c.Set(string(constants.UserIDKey), tc.UserID)
					}
					h.DeleteTransaction(c)
				})
				router.ServeHTTP(w, req)
			}

			var actualResponse map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &actualResponse)
			require.NoError(t, err)

			if tc.ExpectedError != "" {
				require.Contains(t, actualResponse["error"].(string), tc.ExpectedError)
			} else {
				if diff := cmp.Diff(tc.ExpectedResponse, actualResponse); diff != "" {
					t.Errorf("response mismatch (-want +got)\n%s", diff)
				}
			}
			if diff := cmp.Diff(tc.ExpectedStatusCode, w.Code); diff != "" {
				t.Errorf("status code mismatch (-want +got)\n%s", diff)
			}
			mockTxQ.AssertExpectations(t)
		})
	}
}

package transaction_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	htx "github.com/seanhuebl/unity-wealth/handlers/transaction"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	dbmocks "github.com/seanhuebl/unity-wealth/internal/mocks/database"
	"github.com/seanhuebl/unity-wealth/internal/services/transaction"
	"github.com/seanhuebl/unity-wealth/internal/testfixtures"
	"github.com/seanhuebl/unity-wealth/internal/testhelpers"
	"github.com/seanhuebl/unity-wealth/internal/testmodels"
	"github.com/stretchr/testify/mock"
)

func TestDeleteTransaction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []testmodels.DeleteTxTestCase{
		{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testfixtures.NilUserID,
				TxID:             uuid.NewString(),
			},
		},
		{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testfixtures.InvalidUserID,
				TxID:             uuid.NewString(),
			},
		},
		{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testfixtures.InvalidTxID,
				TxID:             "",
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
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			mockTxQ := dbmocks.NewTransactionQuerier(t)
			w := httptest.NewRecorder()
			svc := transaction.NewTransactionService(mockTxQ)
			req := httptest.NewRequest("DELETE", fmt.Sprintf("/transactions/%v", tc.TxID), nil)

			if tc.UserIDErr == nil && tc.TxID != "" {
				mockTxQ.On("DeleteTransactionByID", context.Background(), mock.AnythingOfType("database.DeleteTransactionByIDParams")).Return(tc.TxID, tc.TxErr)
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

			actualResponse := testhelpers.ProcessResponse(w, t)
			testhelpers.CheckHTTPResponse(t, w, tc.ExpectedError, tc.ExpectedStatusCode, tc.ExpectedResponse, actualResponse)

			mockTxQ.AssertExpectations(t)
		})
	}
}

package transaction_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	htx "github.com/seanhuebl/unity-wealth/handlers/transaction"
	handlermocks "github.com/seanhuebl/unity-wealth/internal/mocks/handlers"
	"github.com/seanhuebl/unity-wealth/internal/models"
	"github.com/seanhuebl/unity-wealth/internal/testfixtures"
	"github.com/seanhuebl/unity-wealth/internal/testhelpers"
	"github.com/seanhuebl/unity-wealth/internal/testmodels"
	"github.com/stretchr/testify/mock"
)

func TestGetTxByID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []testmodels.GetTxTestCase{
		{
			BaseHTTPTestCase: testfixtures.NilUserID,
			TxID:             uuid.NewString(),
		},
		{
			BaseHTTPTestCase: testfixtures.InvalidUserID,
			TxID:             uuid.NewString(),
		},
		{
			BaseHTTPTestCase: testfixtures.InvalidTxID,
			TxID:             "INVALID",
		},
		{
			BaseHTTPTestCase: testfixtures.EmptyTxID,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			mockSvc := handlermocks.NewTransactionService(t)
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", fmt.Sprintf("/transactions/%v", tc.TxID), nil)
			h := htx.NewHandler(mockSvc)
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Params = gin.Params{{Key: "id", Value: tc.TxID}}
				testhelpers.CheckForUserIDIssues(tc.Name, tc.UserID, c)
				c.Next()
			})
			router.NoRoute(func(c *gin.Context) {
				c.JSON(http.StatusNotFound, gin.H{
					"data": gin.H{"error": "not found"},
				})
			})
			router.GET("/transactions/:id", h.GetTransactionByID)
			router.ServeHTTP(w, req)
			actualResponse := testhelpers.ProcessResponse(w, t)
			testhelpers.CheckHTTPResponse(t, w, tc.ExpectedError, tc.ExpectedStatusCode, tc.ExpectedResponse, actualResponse)
		})
	}

	txErrTests := []testmodels.GetTxTestCase{
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{

				Name:               "error getting tx",
				UserID:             uuid.New(),
				ExpectedError:      "unable to get transaction",
				ExpectedStatusCode: http.StatusInternalServerError,
				ExpectedResponse: map[string]interface{}{
					"data": map[string]interface{}{
						"error": "unable to get transaction",
					},
				},
			},
			TxID:  uuid.NewString(),
			TxErr: errors.New("error getting transaction"),
		},
		{
			BaseHTTPTestCase: testmodels.BaseHTTPTestCase{

				Name:               "not found",
				UserID:             uuid.New(),
				ExpectedError:      "not found",
				ExpectedStatusCode: http.StatusNotFound,
				ExpectedResponse: map[string]interface{}{
					"data": map[string]interface{}{
						"error": "not found",
					},
				},
			},
			TxID:  uuid.NewString(),
			TxErr: errors.New("transaction not found"),
		},
	}

	for _, tc := range txErrTests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			w := httptest.NewRecorder()
			mockSvc := handlermocks.NewTransactionService(t)

			req := httptest.NewRequest("GET", fmt.Sprintf("/transactions/%v", tc.TxID), nil)

			mockSvc.On("GetTransactionByID", mock.Anything, tc.UserID.String(), tc.TxID).Return((*models.Tx)(nil), tc.TxErr)

			h := htx.NewHandler(mockSvc)

			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Params = gin.Params{{Key: "id", Value: tc.TxID}}
				testhelpers.CheckForUserIDIssues(tc.Name, tc.UserID, c)
				c.Next()
			})

			router.GET("/transactions/:id", h.GetTransactionByID)

			router.ServeHTTP(w, req)
			mockSvc.AssertCalled(t,
				"GetTransactionByID",
				mock.Anything,
				tc.UserID.String(),
				tc.TxID,
			)

			actualResponse := testhelpers.ProcessResponse(w, t)
			testhelpers.CheckHTTPResponse(t, w, tc.ExpectedError, tc.ExpectedStatusCode, tc.ExpectedResponse, actualResponse)
			mockSvc.AssertExpectations(t)
		})
	}
}

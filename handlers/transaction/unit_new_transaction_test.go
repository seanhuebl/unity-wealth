package transaction_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	htx "github.com/seanhuebl/unity-wealth/handlers/transaction"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	handlermocks "github.com/seanhuebl/unity-wealth/internal/mocks/handlers"
	"github.com/seanhuebl/unity-wealth/internal/testfixtures"
	"github.com/seanhuebl/unity-wealth/internal/testhelpers"
	"github.com/seanhuebl/unity-wealth/internal/testmodels"
	"github.com/stretchr/testify/mock"
)

func TestNewTx(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	tests := []testmodels.CreateTxTestCase{
		{
			BaseHTTPTestCase: testfixtures.NilUserID,
			ReqBody:          `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
		},
		{
			BaseHTTPTestCase: testfixtures.InvalidUserID,
			ReqBody:          `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
		},
		{
			BaseHTTPTestCase: testfixtures.InvalidReqBody,
			ReqBody:          `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40`,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			mockSvc := handlermocks.NewTransactionService(t)
			
			t.Cleanup(func() { mockSvc.AssertExpectations(t) })
			req := httptest.NewRequest("POST", "/transactions", bytes.NewBufferString(tc.ReqBody))
			req.Header.Set("Content-Type", "application/json")

			h := htx.NewHandler(mockSvc)

			router := gin.New()
			router.Use(func(c *gin.Context) {
				testhelpers.CheckForUserIDIssues(tc.Name, tc.UserID, c)

				c.Next()
			})

			router.POST("/transactions", h.NewTransaction)
			router.ServeHTTP(w, req)

			actualResponse := testhelpers.ProcessResponse(w, t)

			testhelpers.CheckHTTPResponse(t, w, tc.ExpectedError, tc.ExpectedStatusCode, tc.ExpectedResponse, actualResponse)
		})
	}
	t.Run("failed to create tx", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		mockSvc := handlermocks.NewTransactionService(t)
		expErr := errors.New("failed to create transaction")
		userID := uuid.New()
		t.Cleanup(func() { mockSvc.AssertExpectations(t) })

		req := httptest.NewRequest("POST", "/transactions", bytes.NewBufferString(`{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`))
		req.Header.Set("Content-Type", "application/json")

		mockSvc.On("CreateTransaction", mock.Anything, userID, mock.Anything).Return(nil, expErr)
		h := htx.NewHandler(mockSvc)

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set(string(constants.UserIDKey), userID)
			c.Next()
		})

		router.POST("/transactions", h.NewTransaction)
		router.ServeHTTP(w, req)

		actualResponse := testhelpers.ProcessResponse(w, t)

		testhelpers.CheckHTTPResponse(
			t,
			w,
			expErr.Error(),
			http.StatusInternalServerError,
			map[string]interface{}{
				"data": map[string]interface{}{
					"error": "failed to create transaction",
				},
			},
			actualResponse,
		)
	})
}

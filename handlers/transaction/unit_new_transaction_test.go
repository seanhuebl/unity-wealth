package transaction

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func TestNewTx(t *testing.T) {
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
			reqBody:            `{"date": "2025-03-05", "merchant": "costco", "amount": "125.98", "detailed_category": "40"}`,
			expectedStatusCode: http.StatusCreated,
			expectedResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"date":              "2025-03-05",
					"merchant":          "costco",
					"amount":            "125.98",
					"detailed_category": "40",
				},
			},
		},
	}
}

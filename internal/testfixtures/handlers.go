package testfixtures

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/testmodels"
)

var (
	NilUserID = testmodels.BaseHTTPTestCase{
		Name:               "unauthorized: user ID is uuid.NIL",
		UserID:             uuid.Nil,
		UserIDErr:          errors.New("user ID not found in context"),
		ExpectedError:      "unauthorized",
		ExpectedStatusCode: http.StatusUnauthorized,
		ExpectedResponse: map[string]interface{}{
			"data": map[string]interface{}{
				"error": "unauthorized",
			},
		},
	}

	InvalidUserID = testmodels.BaseHTTPTestCase{
		Name:               "unauthorized: user ID not UUID",
		UserID:             uuid.New(), // overidden in test
		UserIDErr:          errors.New("user ID is not UUID"),
		ExpectedStatusCode: http.StatusUnauthorized,
		ExpectedResponse: map[string]interface{}{
			"data": map[string]interface{}{
				"error": "unauthorized",
			},
		},
	}

	InvalidTxID = testmodels.BaseHTTPTestCase{
		Name:               "invalid txID in req",
		UserID:             uuid.New(),
		ExpectedError:      "invalid id",
		ExpectedStatusCode: http.StatusBadRequest,
		ExpectedResponse: map[string]interface{}{
			"data": map[string]interface{}{
				"error": "invalid id",
			},
		},
	}

	EmptyTxID = testmodels.BaseHTTPTestCase{
		Name:               "empty txID in req",
		UserID:             uuid.New(),
		ExpectedError:      "not found",
		ExpectedStatusCode: http.StatusNotFound,
		ExpectedResponse: map[string]interface{}{
			"data": map[string]interface{}{
				"error": "not found",
			},
		},
	}

	InvalidReqBody = testmodels.BaseHTTPTestCase{
		Name:               "invalid request body",
		UserID:             uuid.New(),
		ExpectedStatusCode: http.StatusBadRequest,
		ExpectedResponse: map[string]interface{}{
			"data": map[string]interface{}{
				"error": "invalid request body",
			},
		},
	}

	NotFound = testmodels.BaseHTTPTestCase{
		Name:               "not found",
		UserID:             uuid.New(),
		ExpectedStatusCode: http.StatusNotFound,
		ExpectedResponse: map[string]interface{}{
			"data": map[string]interface{}{
				"error": "not found",
			},
		},
	}
)

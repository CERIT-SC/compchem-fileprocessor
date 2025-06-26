package active_workflows

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"fi.muni.cz/invenio-file-processor/v2/httpclient"
	"fi.muni.cz/invenio-file-processor/v2/routes/common"
	"github.com/stretchr/testify/assert"
)

func TestHandleError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "ClientError",
			err: &httpclient.ClientError{
				Status:  400,
				Message: "invalid request",
			},
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "Argo could not process request: Error on client side, status: 400, message: invalid request",
		},
		{
			name: "ServerError",
			err: &httpclient.ServerError{
				Status:  500,
				Message: "database connection failed",
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedMsg:    "Argo might currently be unavailable: Error on server side, status: 500, message: database connection failed",
		},
		{
			name:           "GenericError",
			err:            errors.New("unexpected error"),
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "Something went wrong when processing request: unexpected error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)

			handleError(w, r, tt.err)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response common.ErrorResponse
			err := json.NewDecoder(w.Body).Decode(&response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, response.Message)
		})
	}
}

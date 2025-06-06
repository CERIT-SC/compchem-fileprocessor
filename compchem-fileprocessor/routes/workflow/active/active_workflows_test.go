package active_workflows

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"fi.muni.cz/invenio-file-processor/v2/httpclient"
	"fi.muni.cz/invenio-file-processor/v2/routes/common"
	"fi.muni.cz/invenio-file-processor/v2/service"
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

func TestGetRequestParams(t *testing.T) {
	tests := []struct {
		name           string
		recordId       string
		queryParams    map[string]string
		expectError    bool
		expectedMsg    string
		expectedLimit  int
		expectedSkip   int
		expectedStates []service.State
	}{
		{
			name:           "Default values with no params",
			recordId:       "123",
			queryParams:    map[string]string{},
			expectError:    false,
			expectedLimit:  20,
			expectedSkip:   0,
			expectedStates: []service.State{},
		},
		{
			name:     "Custom limit and skip",
			recordId: "456",
			queryParams: map[string]string{
				"limit": "50",
				"skip":  "10",
			},
			expectError:    false,
			expectedLimit:  50,
			expectedSkip:   10,
			expectedStates: []service.State{},
		},
		{
			name:     "Invalid limit (non-numeric)",
			recordId: "789",
			queryParams: map[string]string{
				"limit": "abc",
			},
			expectError: true,
			expectedMsg: `strconv.Atoi: parsing "abc": invalid syntax`,
		},
		{
			name:     "Invalid skip (non-numeric)",
			recordId: "789",
			queryParams: map[string]string{
				"skip": "xyz",
			},
			expectError: true,
			expectedMsg: `strconv.Atoi: parsing "xyz": invalid syntax`,
		},
		{
			name:     "Valid status filter with single state",
			recordId: "123",
			queryParams: map[string]string{
				"status": "(Running)",
			},
			expectError:    false,
			expectedLimit:  20,
			expectedSkip:   0,
			expectedStates: []service.State{service.StateRunning},
		},
		{
			name:     "Valid status filter with multiple states",
			recordId: "123",
			queryParams: map[string]string{
				"status": "(Running,Pending,Succeeded)",
			},
			expectError:   false,
			expectedLimit: 20,
			expectedSkip:  0,
			expectedStates: []service.State{
				service.StateRunning,
				service.StatePending,
				service.StateSucceeded,
			},
		},
		{
			name:     "Invalid status format (missing parentheses)",
			recordId: "123",
			queryParams: map[string]string{
				"status": "Running,Pending",
			},
			expectError: true,
			expectedMsg: "State filter does not match format: (Running, Pending, Error, Succeeded, Failed)",
		},
		{
			name:     "Invalid status format (wrong brackets)",
			recordId: "123",
			queryParams: map[string]string{
				"status": "[Running,Pending]",
			},
			expectError: true,
			expectedMsg: "State filter does not match format: (Running, Pending, Error, Succeeded, Failed)",
		},
		{
			name:     "Invalid state name",
			recordId: "123",
			queryParams: map[string]string{
				"status": "(InvalidState)",
			},
			expectError: true,
			expectedMsg: "Unknown workflow state: ",
		},
		{
			name:     "Mixed valid and invalid states",
			recordId: "123",
			queryParams: map[string]string{
				"status": "(Running,InvalidState,Pending)",
			},
			expectError: true,
			expectedMsg: "Unknown workflow state: ",
		},
		{
			name:     "All parameters combined",
			recordId: "999",
			queryParams: map[string]string{
				"limit":  "100",
				"skip":   "50",
				"status": "(Error,Failed)",
			},
			expectError:   false,
			expectedLimit: 100,
			expectedSkip:  50,
			expectedStates: []service.State{
				service.StateError,
				service.StateFailed,
			},
		},
		{
			name:     "Empty status filter",
			recordId: "123",
			queryParams: map[string]string{
				"status": "",
			},
			expectError:    false,
			expectedLimit:  20,
			expectedSkip:   0,
			expectedStates: []service.State{},
		},
		{
			name:     "Status with spaces",
			recordId: "123",
			queryParams: map[string]string{
				"status": "(Running, Pending, Error)",
			},
			expectError:   false,
			expectedLimit: 20,
			expectedSkip:  0,
			expectedStates: []service.State{
				service.StateRunning,
				service.StatePending,
				service.StateError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request with path value and query params
			req := httptest.NewRequest("GET", "/test/"+tt.recordId, nil)
			req.SetPathValue("recordId", tt.recordId)

			q := url.Values{}
			for k, v := range tt.queryParams {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()

			w := httptest.NewRecorder()

			params, err := getRequestParams(w, req)

			if tt.expectError {
				assert.Error(t, err, "expected error but got none")

				if err != nil {
					var response common.ErrorResponse
					decodeErr := json.NewDecoder(w.Body).Decode(&response)
					assert.NoError(t, decodeErr, "failed to decode error response")

					assert.Equal(t, tt.expectedMsg, response.Message, "unexpected error message")

					assert.Equal(t, http.StatusBadRequest, w.Code, "unexpected status code")
				}
			} else {
				assert.NoError(t, err, "unexpected error")
				assert.NotNil(t, params, "expected params but got nil")

				if params != nil {
					assert.Equal(t, tt.recordId, params.recordId, "unexpected recordId")
					assert.Equal(t, tt.expectedLimit, params.limit, "unexpected limit")
					assert.Equal(t, tt.expectedSkip, params.skip, "unexpected skip")

					assert.Len(t, params.statusFilter, len(tt.expectedStates), "unexpected number of states")
					for i, expectedState := range tt.expectedStates {
						if i < len(params.statusFilter) {
							assert.Equal(t, expectedState, params.statusFilter[i], "unexpected state at index %d", i)
						}
					}
				}
			}
		})
	}
}

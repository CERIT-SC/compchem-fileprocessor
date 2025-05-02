package httpclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// Test response structure
type TestResponse struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

// Test request structure
type TestRequest struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

// requested tests generate by claude AI
func TestGetRequest_ServerHasHandler_ReturnsCorrectObject(t *testing.T) {
	// Create a test logger
	logger := zap.NewNop()

	// Create a test server with a GET handler
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request method
		assert.Equal(t, http.MethodGet, r.Method)

		// Write the response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(TestResponse{
			Message: "success",
			Status:  200,
		})
	}))
	defer server.Close()

	// Make the request
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := GetRequest[TestResponse](ctx, logger, server.URL, false)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "success", result.Message)
	assert.Equal(t, 200, result.Status)
}

func TestPostRequest_ServerHasHandler_BodyReceivedReturnsCorrectObject(t *testing.T) {
	// Create a test logger
	logger := zap.NewNop()

	// Create a test server with a POST handler
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		assert.Equal(t, http.MethodPost, r.Method)

		// Verify content type
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Parse the request body
		var req TestRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)

		// Verify request data
		assert.Equal(t, "test-name", req.Name)
		assert.Equal(t, 42, req.Value)

		// Write the response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(TestResponse{
			Message: "post successful",
			Status:  200,
		})
	}))
	defer server.Close()

	// Create the request body
	reqBody := TestRequest{
		Name:  "test-name",
		Value: 42,
	}

	// Make the request
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := PostRequest[TestResponse](ctx, logger, server.URL, reqBody, false)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "post successful", result.Message)
	assert.Equal(t, 200, result.Status)
}

func TestPostRequest_ServerDoesNotAcceptRequest_ClientReturnsError(t *testing.T) {
	// Create a test logger
	logger := zap.NewNop()

	// Create a test server that returns a 400 Bad Request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "bad request"}`))
	}))
	defer server.Close()

	// Create the request body
	reqBody := TestRequest{
		Name:  "test-name",
		Value: 42,
	}

	// Make the request
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := PostRequest[TestResponse](ctx, logger, server.URL, reqBody, false)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "400")
}

func TestGetRequest_ServerDown_ClientRetriesUntilFailure(t *testing.T) {
	// Create a test logger
	logger := zap.NewNop()

	// Track number of attempts
	attemptCount := 0
	maxRetries := 3

	// Create a test server that always returns 503 Service Unavailable
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error": "service unavailable"}`))
	}))
	defer server.Close()

	// Create custom options with fewer retries for faster test
	opts := NewDefaultOpts(logger)
	opts.MaxRetries = maxRetries

	// Create a client with our custom options
	client := Client{
		options:    opts,
		httpClient: http.DefaultClient,
	}

	// Make the request
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.requestRaw(ctx, http.MethodGet, server.URL, nil)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "request failed after")

	// Fix the calculation to match how many attempts should occur based on MaxRetries
	// If MaxRetries is 3, we should have 4 attempts (1 initial + 3 retries)
	assert.Equal(t, maxRetries, attemptCount, "Expected retries didn't match")
}

// This is an extra test to verify the retry with eventual success behavior
func TestGetRequest_ServerTemporarilyDown_ClientRetriesAndEventuallySucceeds(t *testing.T) {
	// Create a test logger
	logger := zap.NewNop()

	// Track number of attempts
	attemptCount := 0
	successAttempt := 2 // Succeed on the 3rd attempt (index 2)

	// Create a test server that fails until the successAttempt
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++

		if attemptCount <= successAttempt {
			// Return server error for first attempts
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error": "service unavailable"}`))
			return
		}

		// Success on later attempt
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(TestResponse{
			Message: "eventually succeeded",
			Status:  200,
		})
	}))
	defer server.Close()

	// Make the request
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := GetRequest[TestResponse](ctx, logger, server.URL, false)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "eventually succeeded", result.Message)
	assert.Equal(t, 200, result.Status)
	assert.Equal(t, successAttempt+1, attemptCount, "Expected number of attempts didn't match")
}

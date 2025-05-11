package available

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fi.muni.cz/invenio-file-processor/v2/api/availabledtos"
	"fi.muni.cz/invenio-file-processor/v2/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestAvailableWorkflowsHandler_ValidBody_CorrectResponseReturned(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()
	configs := []config.WorkflowConfig{
		{
			Name:     "count-words",
			Filetype: "text/plain",
			ProcessingTemplates: []config.ProcessingTemplate{
				{
					Name:     "count-words",
					Template: "count-words-template",
				},
			},
		},
		{
			Name:     "downsizeto480p",
			Filetype: "image/jpeg",
			ProcessingTemplates: []config.ProcessingTemplate{
				{
					Name:     "downsizepng",
					Template: "downsizepng-template",
				},
				{
					Name:     "downsizejpeg",
					Template: "downsizejpeg-template",
				},
			},
		},
	}

	reqBody := availabledtos.AvailableWorkflowsRequest{
		Files: []availabledtos.KeyAndType{
			{
				FileKey:  "test1.txt",
				Mimetype: "text/plain",
			},
			{
				FileKey:  "test.jpeg",
				Mimetype: "image/jpeg",
			},
		},
	}
	jsonBody, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/workflows/available", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler := AvailableWorkflowsHandler(ctx, logger, configs)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response availabledtos.AvailableWorkflowsResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Len(t, response.Workflows, 2)

	for _, workflow := range response.Workflows {
		if workflow.Mimetype == "text/plain" {
			assert.Contains(t, workflow.Files, "test1.txt")
		} else if workflow.Mimetype == "image/jpeg" {
			assert.Contains(t, workflow.Files, "test.jpeg")
		} else {
			t.Fatalf("Unexpected mimetype in response: %s", workflow.Mimetype)
		}
	}
}

func TestAvailableWorkflowsHandler_ValidBodyNoConfigs_EmtpyOKResponse(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()
	configs := []config.WorkflowConfig{
		{
			Name:     "count-words",
			Filetype: "text/plain",
			ProcessingTemplates: []config.ProcessingTemplate{
				{
					Name:     "count-words",
					Template: "count-words-template",
				},
			},
		},
	}

	reqBody := availabledtos.AvailableWorkflowsRequest{
		Files: []availabledtos.KeyAndType{},
	}
	jsonBody, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/workflows/available", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler := AvailableWorkflowsHandler(ctx, logger, configs)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response availabledtos.AvailableWorkflowsResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Len(t, response.Workflows, 0)
}

func TestAvailbleWorkflowsHandler_InvalidBody_StatusBadRequest(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()
	configs := []config.WorkflowConfig{}

	testCases := []struct {
		name     string
		reqBody  availabledtos.AvailableWorkflowsRequest
		expected string
	}{
		{
			name: "Missing FileKey",
			reqBody: availabledtos.AvailableWorkflowsRequest{
				Files: []availabledtos.KeyAndType{
					{
						FileKey:  "",
						Mimetype: "text/plain",
					},
				},
			},
			expected: "missing file_key at: 0",
		},
		{
			name: "Missing Mimetype",
			reqBody: availabledtos.AvailableWorkflowsRequest{
				Files: []availabledtos.KeyAndType{
					{
						FileKey:  "test.txt",
						Mimetype: "",
					},
				},
			},
			expected: "missing mimetype at: 0",
		},
		{
			name: "Multiple Missing Fields",
			reqBody: availabledtos.AvailableWorkflowsRequest{
				Files: []availabledtos.KeyAndType{
					{
						FileKey:  "",
						Mimetype: "",
					},
					{
						FileKey:  "test2.txt",
						Mimetype: "",
					},
				},
			},
			expected: "missing file_key at: 0, missing mimetype at: 0, missing mimetype at: 1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonBody, err := json.Marshal(tc.reqBody)
			assert.NoError(t, err)

			req := httptest.NewRequest(
				http.MethodPost,
				"/workflows/available",
				bytes.NewBuffer(jsonBody),
			)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler := AvailableWorkflowsHandler(ctx, logger, configs)
			handler.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, w.Body.String(), tc.expected)
		})
	}
}

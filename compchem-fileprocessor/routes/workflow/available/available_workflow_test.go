package available

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fi.muni.cz/invenio-file-processor/v2/config"
	"fi.muni.cz/invenio-file-processor/v2/services"
	"fi.muni.cz/invenio-file-processor/v2/services/list_workflows"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestAvailableWorkflowsHandler_ValidBody_CorrectResponseReturned(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()
	configs := []config.WorkflowConfig{
		{
			Name:     "count-words",
			Mimetype: "text/plain",
			ProcessingTemplates: []config.ProcessingTemplate{
				{
					Name:     "count-words",
					Template: "count-words-template",
				},
			},
		},
		{
			Name:     "downsizeto480p",
			Mimetype: "image/jpeg",
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

	reqBody := list_workflows.AvailableWorkflowsRequest{
		Files: []services.File{
			{
				FileName: "test1.txt",
				Mimetype: "text/plain",
			},
			{
				FileName: "test.jpeg",
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

	var response list_workflows.AvailableWorkflowsResponse
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
			Mimetype: "text/plain",
			ProcessingTemplates: []config.ProcessingTemplate{
				{
					Name:     "count-words",
					Template: "count-words-template",
				},
			},
		},
	}

	reqBody := list_workflows.AvailableWorkflowsRequest{
		Files: []services.File{},
	}
	jsonBody, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/workflows/available", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler := AvailableWorkflowsHandler(ctx, logger, configs)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response list_workflows.AvailableWorkflowsResponse
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
		reqBody  list_workflows.AvailableWorkflowsRequest
		expected string
	}{
		{
			name: "Missing FileName",
			reqBody: list_workflows.AvailableWorkflowsRequest{
				Files: []services.File{
					{
						FileName: "",
						Mimetype: "text/plain",
					},
				},
			},
			expected: "missing filename at: 0",
		},
		{
			name: "Missing Mimetype",
			reqBody: list_workflows.AvailableWorkflowsRequest{
				Files: []services.File{
					{
						FileName: "test.txt",
						Mimetype: "",
					},
				},
			},
			expected: "missing mimetype at: 0",
		},
		{
			name: "Multiple Missing Fields",
			reqBody: list_workflows.AvailableWorkflowsRequest{
				Files: []services.File{
					{
						FileName: "",
						Mimetype: "",
					},
					{
						FileName: "test2.txt",
						Mimetype: "",
					},
				},
			},
			expected: "missing filename at: 0, missing mimetype at: 0, missing mimetype at: 1",
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

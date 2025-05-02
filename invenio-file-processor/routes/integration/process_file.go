package integration

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"fi.muni.cz/invenio-file-processor/v2/argointegration"
	"fi.muni.cz/invenio-file-processor/v2/config"
	"fi.muni.cz/invenio-file-processor/v2/jsonapi"
	"go.uber.org/zap"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

type requestBody struct {
	RecordId string `json:"recordId"`
	FileName string `json:"fileName"`
	FileType string `json:"fileType"` // TODO: can invenio figure this out or should we?
}

func CommitedFileHandler(
	ctx context.Context,
	logger *zap.Logger,
	argoUrl string,
	baseUrl string,
	configs []config.WorkflowConfig,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}

		reqBody, err := getRequestBody(w, r)
		if err != nil {
			logger.Error("Requst body invalid", zap.Error(err))
			return
		}

		err = argointegration.ProcessCommittedFile(
			ctx,
			logger,
			argoUrl,
			baseUrl,
			reqBody.RecordId,
			reqBody.FileName,
			reqBody.FileType,
			configs,
		)
		if err != nil {
			logger.Error("Failed to submit file for processing", zap.Error(err))
			jsonapi.Encode(w, r, http.StatusInternalServerError, ErrorResponse{
				Message: "Failed to submit workflow to argo",
			})
			return
		}

		logger.Info(
			"File successfully submitted for processing",
			zap.String("recordId", reqBody.RecordId),
			zap.String("filename", reqBody.FileName),
		)
		w.WriteHeader(http.StatusOK)
	})
}

func getRequestBody(w http.ResponseWriter, r *http.Request) (*requestBody, error) {
	reqBody, err := jsonapi.Decode[requestBody](r)
	if err != nil {
		jsonapi.Encode(w, r, 400, ErrorResponse{
			Message: "Failed to decode request for processing",
		})
		return nil, fmt.Errorf("Decode error")
	}

	if err := validateBody(reqBody); err != nil {
		jsonapi.Encode(w, r, 400, ErrorResponse{
			Message: "Invalid request body, missing: " + err.Error(),
		})
		return nil, fmt.Errorf("Validate erorr: %v", err)
	}

	return &reqBody, nil
}

func validateBody(body requestBody) error {
	var errors []string

	if body.FileName == "" {
		errors = append(errors, "fileName")
	}

	if body.FileType == "" {
		errors = append(errors, "fileType")
	}

	if body.RecordId == "" {
		errors = append(errors, "recordId")
	}

	if len(errors) > 0 {
		return fmt.Errorf("Missing attributes: %s", strings.Join(errors, ", "))
	}

	return nil
}

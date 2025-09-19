package start_workflow_route

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"fi.muni.cz/invenio-file-processor/v2/config"
	"fi.muni.cz/invenio-file-processor/v2/jsonapi"
	"fi.muni.cz/invenio-file-processor/v2/routes/common"
	startworkflow_service "fi.muni.cz/invenio-file-processor/v2/services/start_workflow"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type startAllRequestBody struct {
	RecordId string                       `json:"recordId"`
	Files    []startworkflow_service.File `json:"files"`
}

func PostAllWorkflowsHandler(
	ctx context.Context,
	logger *zap.Logger,
	pool *pgxpool.Pool,
	argoUrl string,
	baseUrl string,
	configs []config.WorkflowConfig,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, err := common.GetValidRequestBody(w, r, validateStartAllBody)
		if err != nil {
			logger.Error("Requst body invalid", zap.Error(err))
			return
		}

		err = startworkflow_service.StartAllWorkflows(
			ctx,
			logger,
			pool,
			argoUrl,
			baseUrl,
			reqBody.RecordId,
			reqBody.Files,
			configs,
		)
		if err != nil {
			logger.Error("Failed to submit file for processing", zap.Error(err))
			jsonapi.Encode(w, r, http.StatusInternalServerError, common.ErrorResponse{
				Message: "Failed to submit workflow to argo",
			})
			return
		}

		logger.Info(
			"File successfully submitted for processing",
			zap.String("recordId", reqBody.RecordId),
		)
		w.WriteHeader(http.StatusCreated)
	})
}

func validateStartAllBody(body *startAllRequestBody) error {
	var errors []string

	if body.RecordId == "" {
		errors = append(errors, "recordId")
	}

	if len(body.Files) > 0 {
		validateFiles(body.Files, errors)
	} else {
		errors = append(errors, "files")
	}

	if len(errors) > 0 {
		return fmt.Errorf("Missing attributes: %s", strings.Join(errors, ", "))
	}

	return nil
}

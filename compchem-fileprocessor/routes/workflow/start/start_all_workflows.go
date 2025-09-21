package start_workflow_route

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"fi.muni.cz/invenio-file-processor/v2/config"
	"fi.muni.cz/invenio-file-processor/v2/jsonapi"
	"fi.muni.cz/invenio-file-processor/v2/routes/common"
	"fi.muni.cz/invenio-file-processor/v2/services"
	startworkflow_service "fi.muni.cz/invenio-file-processor/v2/services/start_workflow"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type startAllRequestBody struct {
	Files []services.File `json:"files"`
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
		recordId := r.PathValue("recordId")
		reqBody, err := common.GetValidRequestBody(w, r, validateStartAllBody)
		if err != nil {
			logger.Error("Requst body invalid", zap.Error(err))
			return
		}

		response, err := startworkflow_service.StartAllWorkflows(
			ctx,
			logger,
			pool,
			argoUrl,
			baseUrl,
			recordId,
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

		err = jsonapi.Encode(w, r, http.StatusCreated, response)
		if err != nil {
			logger.Error(
				"Failed to Encode response for post all workflows handler",
				zap.Any("response", response),
				zap.Error(err),
			)
		}
	})
}

func validateStartAllBody(body *startAllRequestBody) error {
	var errors []string

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

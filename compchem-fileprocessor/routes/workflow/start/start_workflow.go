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

type startRequestBody struct {
	Name  string          `json:"name"`
	Files []services.File `json:"files"`
}

func PostWorkflowHandler(
	ctx context.Context,
	logger *zap.Logger,
	pool *pgxpool.Pool,
	argoUrl string,
	baseUrl string,
	configs []config.WorkflowConfig,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recordId := r.PathValue("recordId")
		reqBody, err := common.GetValidRequestBody(w, r, validateStartBody)
		if err != nil {
			logger.Error("Requst body invalid", zap.Error(err))
			return
		}

		err = startworkflow_service.StartWorkflow(
			ctx,
			logger,
			pool,
			argoUrl,
			baseUrl,
			reqBody.Name,
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

		logger.Info(
			"File successfully submitted for processing",
			zap.String("recordId", recordId),
		)
		w.WriteHeader(http.StatusCreated)
	})
}

func validateStartBody(body *startRequestBody) error {
	var errors []string

	if body.Name == "" {
		errors = append(errors, "name")
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

package start_workflow_route

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"fi.muni.cz/invenio-file-processor/v2/config"
	"fi.muni.cz/invenio-file-processor/v2/jsonapi"
	"fi.muni.cz/invenio-file-processor/v2/routes/common"
	"fi.muni.cz/invenio-file-processor/v2/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type requestBody struct {
	Files    []service.File `json:"files"`
	RecordId string         `json:"recordId"`
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
		reqBody, err := common.GetValidRequestBody(w, r, validateBody)
		if err != nil {
			logger.Error("Requst body invalid", zap.Error(err))
			return
		}

		err = service.StartWorkflow(
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

func validateBody(body *requestBody) error {
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

func validateFiles(files []service.File, errors []string) {
	mimetype := files[0].Mimetype

	for index, file := range files {
		if file.FileName == "" {
			errors = append(errors, fmt.Sprintf("fileName-%d", index))
		}

		if file.Mimetype == "" || file.Mimetype != mimetype {
			errors = append(errors, fmt.Sprintf("mimetype-%d", index))
		}
	}
}

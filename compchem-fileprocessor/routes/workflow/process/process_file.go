package process

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
	RecordId string `json:"recordId"`
	FileName string `json:"fileName"`
	Mimetype string `json:"mimetype"`
}

func CommitedFileHandler(
	ctx context.Context,
	logger *zap.Logger,
	pool *pgxpool.Pool,
	argoUrl string,
	baseUrl string,
	configs []config.WorkflowConfig,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := common.ValidateMethod(w, r, http.MethodPost)
		if err != nil {
			http.Error(w, err.Error(), http.StatusMethodNotAllowed)
			return
		}

		reqBody, err := common.GetValidRequestBody(w, r, validateBody)
		if err != nil {
			logger.Error("Requst body invalid", zap.Error(err))
			return
		}

		file, err := service.ProcessCommittedFile(
			ctx,
			logger,
			pool,
			argoUrl,
			baseUrl,
			reqBody.RecordId,
			reqBody.FileName,
			reqBody.Mimetype,
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
			zap.String("filename", reqBody.FileName),
		)
		jsonapi.Encode(w, r, http.StatusCreated, file)
	})
}

func validateBody(body *requestBody) error {
	var errors []string

	if body.FileName == "" {
		errors = append(errors, "fileName")
	}

	if body.Mimetype == "" {
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

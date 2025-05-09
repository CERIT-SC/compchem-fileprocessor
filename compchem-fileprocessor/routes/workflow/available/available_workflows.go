package available

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"fi.muni.cz/invenio-file-processor/v2/api/availabledtos"
	"fi.muni.cz/invenio-file-processor/v2/config"
	"fi.muni.cz/invenio-file-processor/v2/routes/common"
	"fi.muni.cz/invenio-file-processor/v2/service"
	"go.uber.org/zap"
)

func AvailableWorkflowsHandler(
	ctx context.Context,
	logger *zap.Logger,
	configs []config.WorkflowConfig,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received request for available workflows")
		_, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
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

		if len(reqBody.Files) == 0 {
			common.EncodeResponse(
				w,
				r,
				http.StatusOK,
				availabledtos.AvailableWorkflowsResponse{
					Workflows: []availabledtos.AvailableWorkflow{},
				},
			)
			return
		}

		response := service.AvailableWorkflows(logger, reqBody, configs)

		common.EncodeResponse(w, r, http.StatusOK, response)

		return
	})
}

func validateBody(req *availabledtos.AvailableWorkflowsRequest) error {
	errors := []string{}

	for i, file := range req.Files {
		if file.FileKey == "" {
			errors = append(errors, fmt.Sprintf("missing file_key at: %d", i))
		}
		if file.Mimetype == "" {
			errors = append(errors, fmt.Sprintf("missing mimetype at: %d", i))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors: %s ", strings.Join(errors, ", "))
	}

	return nil
}

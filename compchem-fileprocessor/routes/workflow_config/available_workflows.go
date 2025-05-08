package workflowconfig

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"fi.muni.cz/invenio-file-processor/v2/config"
	"fi.muni.cz/invenio-file-processor/v2/jsonapi"
	"fi.muni.cz/invenio-file-processor/v2/routes/common"
	"go.uber.org/zap"
)

type AvailableWorkflowsRequest struct {
	Files []keyAndType `json:"files"`
}

type keyAndType struct {
	FileKey  string `json:"key"`
	Mimetype string `json:"mimetype"`
}

type AvailableWorkflowsResponse struct {
	Available []availableWorkflow
}

type availableWorkflow struct {
	Mimetype string `json:"mimetype"`
	Files    []File `json:"files"`
}

type File struct {
	Key string `json:"key"`
}

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

		reqBody, err := common.GetRequestBody[AvailableWorkflowsRequest](w, r, validateBody)
		if err != nil {
			logger.Error("Requst body invalid", zap.Error(err))
			return
		}

		if len(reqBody.Files) == 0 {
			jsonapi.Encode(w, r, 200, AvailableWorkflowsResponse{Available: []availableWorkflow{}})
			return
		}

		return
	})
}

func validateBody(req *AvailableWorkflowsRequest) error {
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

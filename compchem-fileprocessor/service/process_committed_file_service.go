package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"fi.muni.cz/invenio-file-processor/v2/api/argodtos"
	"fi.muni.cz/invenio-file-processor/v2/config"
	"fi.muni.cz/invenio-file-processor/v2/httpclient"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// for now one process for easch file type
func ProcessCommittedFile(
	ctx context.Context,
	logger *zap.Logger,
	argoUrl string,
	baseUrl string,
	recordId string,
	fileName string,
	filetype string,
	configs []config.WorkflowConfig,
) error {
	conf, err := findWorkflowConfig(configs, filetype)
	if err != nil {
		return err
	}

	workflow := argodtos.BuildWorkflow(
		*conf,
		baseUrl,
		conf.Name,
		uuid.New().String(), // TODO: replace with sequential numbering
		recordId,
		[]string{fileName},
	)

	err = submitWorkflow(ctx, logger, argoUrl, workflow)
	if err != nil {
		logger.Error(
			"failed to submit workflow",
			zap.String("fileName", fileName),
			zap.String("recordId", recordId),
			zap.String("filetype", filetype),
			zap.Any("config", conf),
		)
		return err
	}

	return nil
}

func findWorkflowConfig(
	configs []config.WorkflowConfig,
	filetype string,
) (*config.WorkflowConfig, error) {
	for _, conf := range configs {
		if conf.Filetype == filetype {
			return &conf, nil
		}
	}

	return nil, errors.New("No configuration found for: " + filetype)
}

func submitWorkflow(
	ctx context.Context,
	logger *zap.Logger,
	argoUrl string,
	workflow *argodtos.Workflow,
) error {
	url := buildWorkflowUrl("argo", argoUrl)
	logger.Info(
		"Submitting workflow to argo",
		zap.String("workflow-name", workflow.Metadata.Name),
		zap.String("url", url),
	)

	_, err := httpclient.PostRequest[any](ctx, logger, url, &argodtos.WorkflowWrapper{
		Workflow: *workflow,
	}, true)
	if err != nil {
		logger.Error("failed to submit workflow", zap.Error(err))
		return err
	}
	return nil
}

func buildWorkflowUrl(namespace string, argoUrl string, more ...string) string {
	if len(more) == 0 {
		return fmt.Sprintf("%s/api/v1/workflows/%s", argoUrl, namespace)
	}
	return fmt.Sprintf("%s/api/v1/workflows/%s/%s", argoUrl, namespace, strings.Join(more, "/"))
}

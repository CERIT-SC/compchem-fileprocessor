package service

import (
	"fi.muni.cz/invenio-file-processor/v2/config"
	workflowconfig "fi.muni.cz/invenio-file-processor/v2/routes/workflow_config"
	"go.uber.org/zap"
)

func AvailableWorkflows(
	logger *zap.Logger,
	request workflowconfig.AvailableWorkflowsRequest,
	configs []config.WorkflowConfig,
) workflowconfig.AvailableWorkflowsResponse {
}

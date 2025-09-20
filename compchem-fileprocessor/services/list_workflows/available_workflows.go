package list_workflows

import (
	"fi.muni.cz/invenio-file-processor/v2/config"
	"fi.muni.cz/invenio-file-processor/v2/services"
	"go.uber.org/zap"
)

type AvailableWorkflowsRequest struct {
	Files []services.File `json:"files"`
}

type AvailableWorkflowsResponse struct {
	Workflows []AvailableWorkflow `json:"workflows"`
}

type AvailableWorkflow struct {
	Name     string   `json:"name"`
	Mimetype string   `json:"mimetype"`
	Files    []string `json:"files"`
}

func AvailableWorkflows(
	logger *zap.Logger,
	request *AvailableWorkflowsRequest,
	configs []config.WorkflowConfig,
) *AvailableWorkflowsResponse {
	fileMap := convertRequestToMap(request)

	return convertMapToAvailableWorkflows(fileMap, configs)
}

func convertMapToAvailableWorkflows(
	mimeTypeMap map[string][]string,
	configs []config.WorkflowConfig,
) *AvailableWorkflowsResponse {
	workflows := []AvailableWorkflow{}
	for _, workflow := range configs {
		if eligibleFiles, isPresent := mimeTypeMap[workflow.Filetype]; isPresent {
			workflows = append(workflows, AvailableWorkflow{
				Name:     workflow.Name,
				Mimetype: workflow.Filetype,
				Files:    eligibleFiles,
			})
		}
	}

	return &AvailableWorkflowsResponse{
		Workflows: workflows,
	}
}

func convertRequestToMap(request *AvailableWorkflowsRequest) map[string][]string {
	result := make(map[string][]string)

	for _, file := range request.Files {
		if filesForType, isPresent := result[file.Mimetype]; isPresent {
			result[file.Mimetype] = append(filesForType, file.FileName)
		} else {
			result[file.Mimetype] = []string{file.FileName}
		}
	}

	return result
}

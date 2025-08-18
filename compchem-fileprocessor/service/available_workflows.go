package service

import (
	"fi.muni.cz/invenio-file-processor/v2/api/availabledtos"
	"fi.muni.cz/invenio-file-processor/v2/config"
	"go.uber.org/zap"
)

func AvailableWorkflows(
	logger *zap.Logger,
	request *availabledtos.AvailableWorkflowsRequest,
	configs []config.WorkflowConfig,
) *availabledtos.AvailableWorkflowsResponse {
	fileMap := convertRequestToMap(request)

	return convertMapToAvailableWorkflows(fileMap, configs)
}

func convertMapToAvailableWorkflows(
	mimeTypeMap map[string][]string,
	configs []config.WorkflowConfig,
) *availabledtos.AvailableWorkflowsResponse {
	workflows := []availabledtos.AvailableWorkflow{}
	for _, workflow := range configs {
		if eligibleFiles, isPresent := mimeTypeMap[workflow.Filetype]; isPresent {
			workflows = append(workflows, availabledtos.AvailableWorkflow{
				Name:     workflow.Name,
				Mimetype: workflow.Filetype,
				Files:    eligibleFiles,
			})
		}
	}

	return &availabledtos.AvailableWorkflowsResponse{
		Workflows: workflows,
	}
}

func convertRequestToMap(request *availabledtos.AvailableWorkflowsRequest) map[string][]string {
	result := make(map[string][]string)

	for _, file := range request.Files {
		if filesForType, isPresent := result[file.Mimetype]; isPresent {
			result[file.Mimetype] = append(filesForType, file.FileKey)
		} else {
			result[file.Mimetype] = []string{file.FileKey}
		}
	}

	return result
}

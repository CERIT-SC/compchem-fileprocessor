package requests

import "fmt"

const WriteFilesTemplate = "write-files-%s-%s"

func NewWriteWorkflow(
	recordId string,
	workflowId string,
	previousTask string,
) *Task {
	return &Task{
		Name:         fmt.Sprintf(WriteFilesTemplate, recordId, workflowId),
		Dependencies: []string{previousTask},
		TemplateReference: TemplateReference{
			Name:     "write-files-template",
			Template: "upload-files",
		},
		Arguments: ParametersAndArtifacts{
			Parameters: []Parameter{
				{
					Name:  "base-url",
					Value: "{{workflow.parameters.base-url}}",
				},
				{
					Name:  "record-id",
					Value: "{{workflow.parameters.record-id}}",
				},
			},
			Artifacts: []Artifact{
				{
					Name: "input-files",
					From: fmt.Sprintf(
						"{{tasks.%s.outputs.artifacts.output-files}}", previousTask,
					),
				},
			},
		},
	}
}

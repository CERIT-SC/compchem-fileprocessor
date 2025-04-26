package requests

import "fmt"

const WriteFilesTemplate = "write-files-%s-%s"

func NewWriteWorkflow(
	name string,
	deps string,
	predecessor string,
	recordId string,
	workflowId string,
) *Task {
	return &Task{
		Name:         fmt.Sprintf(WriteFilesTemplate, recordId, workflowId),
		Dependencies: deps,
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
					Name: "downloaded-files",
					From: fmt.Sprintf(
						"{{tasks."+ReadFilesTemplate+".outputs.artifacts.downloaded-files}}",
						recordId,
						workflowId,
					),
				},
			},
		},
	}
}

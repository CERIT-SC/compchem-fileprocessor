package argodtos

import "fmt"

const WriteFilesTemplate = "write-files-%s-%s-%d"

func NewWriteWorkflow(
	recordId string,
	workflowId uint64,
	previousTaskFullName string,
	previousTaskTemplateName string,
) *Task {
	return &Task{
		Name: fmt.Sprintf(
			WriteFilesTemplate,
			previousTaskTemplateName,
			recordId,
			workflowId,
		),
		Dependencies: []string{previousTaskFullName},
		TemplateReference: TemplateReference{
			Name:     "write-files-template",
			Template: "write-files",
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
						"{{tasks.%s.outputs.artifacts.output-files}}", previousTaskFullName,
					),
				},
			},
		},
	}
}

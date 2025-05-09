package argodtos

import "fmt"

const ReadFilesTemplate = "read-files-%s-%s"

func NewReadFilesWorkflow(
	recordId string,
	workflowId string,
) *Task {
	return &Task{
		Name:         fmt.Sprintf(ReadFilesTemplate, recordId, workflowId),
		Dependencies: []string{},
		TemplateReference: TemplateReference{
			Name:     "read-files-template",
			Template: "read-files",
		},
		Arguments: ParametersAndArtifacts{
			Artifacts: []Artifact{},
			Parameters: []Parameter{
				{
					Name:  "base-url",
					Value: "{{workflow.parameters.base-url}}",
				},
				{
					Name:  "record-id",
					Value: "{{workflow.parameters.record-id}}",
				},
				{
					Name:  "file-ids",
					Value: "{{workflow.parameters.file-ids}}",
				},
			},
		},
	}
}

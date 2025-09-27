package argodtos

import "fmt"

const writeFilesTemplate = "write-files-%s-%s-%d"

func newWriteWorkflow(
	recordId string,
	workflowId uint64,
	previousTaskFullName string,
	previousTaskTemplateName string,
	workflowFullName string,
) *Task {
	return &Task{
		Name: fmt.Sprintf(
			writeFilesTemplate,
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
				{
					Name:  "secret-key",
					Value: "{{workflow.parameters.secret-key}}",
				},
				{
					Name:  "workflow-name",
					Value: workflowFullName,
				},
				{
					Name:  "task-discriminator",
					Value: previousTaskTemplateName,
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

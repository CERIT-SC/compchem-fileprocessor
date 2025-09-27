package argodtos

import "fmt"

const deleteTokenTemplate = "delete-token-%s-%d"

func newDeleteWorkflow(
	recordId string,
	workflowId uint64,
	workflowFullName string,
	secretKey string,
	previousTasks []string,
) *Task {
	return &Task{
		Name:         fmt.Sprintf(deleteTokenTemplate, recordId, workflowId),
		Dependencies: previousTasks,
		TemplateReference: TemplateReference{
			Name:     "delete-token-template",
			Template: "delete-token",
		},
		Arguments: ParametersAndArtifacts{
			Artifacts: []Artifact{},
			Parameters: []Parameter{
				{
					Name:  "base-url",
					Value: "{{workflow.parameters.base-url}}",
				},
				{
					Name:  "workflow-name",
					Value: workflowFullName,
				},
				{
					Name:  "secret-key",
					Value: "{{workflow.parameters.secret-key}}",
				},
			},
		},
	}
}

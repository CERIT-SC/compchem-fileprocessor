package argodtos

import "fmt"

const deleteContextTemplate = "delete-context-%s-%d"

func newDeleteWorkflow(
	recordId string,
	workflowId uint64,
	workflowFullName string,
	previousTasks []string,
) *Task {
	return &Task{
		Name:         fmt.Sprintf(deleteContextTemplate, recordId, workflowId),
		Dependencies: previousTasks,
		TemplateReference: TemplateReference{
			Name:     "delete-context-template",
			Template: "delete-context",
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

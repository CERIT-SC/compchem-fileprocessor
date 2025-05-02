package requests

import "fmt"

func NewProcessingStep(
	recordId string,
	workflowId string,
	previousTask string,
	templateRef *TemplateReference,
) *Task {
	template := templateRef.Template

	return &Task{
		Name:              fmt.Sprintf(template+"%s-%s", recordId, workflowId),
		Dependencies:      []string{previousTask},
		TemplateReference: *templateRef,
		Arguments: ParametersAndArtifacts{
			Artifacts: []Artifact{
				{
					Name: "input-files",
					From: fmt.Sprintf("{{tasks.%s.outputs.artifacts.output-files}}", previousTask),
				},
			},
			Parameters: []Parameter{},
		},
	}
}

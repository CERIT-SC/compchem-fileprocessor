package argodtos

import "fmt"

func NewProcessingStep(
	recordId string,
	workflowId uint64,
	previousTask string,
	templateRef *TemplateReference,
) *Task {
	template := templateRef.Template

	return &Task{
		Name:              fmt.Sprintf(template+"-%s-%d", recordId, workflowId),
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

package requests

import (
	"fmt"
	"strings"
)

type Workflow struct {
	ApiVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
	Spec       Spec     `yaml:"spec"`
}

type Metadata struct {
	Name string `yaml:"name"`
}

type Spec struct {
	Entrypoint string     `yaml:"entrypoint"`
	Arguments  Arguments  `yaml:"arguments"`
	Templates  []Template `yaml:"templates"`
}

type Arguments struct {
	Parameters []Parameter `yaml:"parameters"`
}

type Template struct {
	Name string `yaml:"name"`
	Dag  Dag    `yaml:"dag"`
}

type Dag struct {
	Tasks []Task `yaml:"tasks"`
}

func constructWorkflowName(workflowName string, recordId string, workflowId string) string {
	return fmt.Sprintf("%s-%s-%s", workflowName, recordId, workflowId)
}

func NewWorkflow(workflowName string,
	recordId string,
	workflowId string,
	baseUrl string,
	fileIds []string,
	processingTasks []Task,
) *Workflow {
	entrypoint := constructWorkflowName(workflowName, recordId, workflowId)

	return &Workflow{
		ApiVersion: "argoproj.io/v1alphat1",
		Kind:       "Workflow",
		Metadata: Metadata{
			Name: entrypoint,
		},
		Spec: Spec{
			Entrypoint: entrypoint,
			Arguments: Arguments{
				Parameters: []Parameter{
					{
						Name:  "base-url",
						Value: baseUrl,
					},
					{
						Name:  "file-ids",
						Value: strings.Join(fileIds, " "),
					},
					{
						Name:  "record-id",
						Value: recordId,
					},
				},
			},
			Templates: []Template{
				{
					Name: entrypoint,
					Dag: Dag{
						Tasks: processingTasks,
					},
				},
			},
		},
	}
}

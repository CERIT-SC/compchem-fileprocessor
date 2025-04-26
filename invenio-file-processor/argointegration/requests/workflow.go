package requests

import (
	"fmt"
	"strings"

	"fi.muni.cz/invenio-file-processor/v2/config"
)

type Workflow struct {
	ApiVersion string   `json:"apiVersion"`
	Kind       string   `json:"kind"`
	Metadata   Metadata `json:"metadata"`
	Spec       Spec     `json:"spec"`
}

type Metadata struct {
	Name string `json:"name"`
}

type Spec struct {
	Entrypoint string     `json:"entrypoint"`
	Arguments  Arguments  `json:"arguments"`
	Templates  []Template `json:"templates"`
}

type Arguments struct {
	Parameters []Parameter `json:"parameters"`
}

type Template struct {
	Name string `json:"name"`
	Dag  Dag    `json:"dag"`
}

type Dag struct {
	Tasks []*Task `json:"tasks"`
}

func constructWorkflowName(workflowName string, recordId string, workflowId string) string {
	return fmt.Sprintf("%s-%s-%s", workflowName, recordId, workflowId)
}

func BuildWorkflow(conf config.WorkflowConfig, baseUrl string, workflowName string, workflowId string, recordId string) *Workflow {
	tasks := constructLinearDag(conf.ProcessingTemplates, recordId, workflowId)

	return newWorkflow(workflowName, recordId, workflowId, baseUrl, []string{}, tasks)
}

func newWorkflow(workflowName string,
	recordId string,
	workflowId string,
	baseUrl string,
	fileIds []string,
	processingTasks []*Task,
) *Workflow {
	entrypoint := constructWorkflowName(workflowName, recordId, workflowId)

	return &Workflow{
		ApiVersion: "argoproj.io/v1alpha1",
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

func constructLinearDag(conf []config.ProcessingTemplate, recordId string, workflowId string) []*Task {
	result := []*Task{}

	readStep := NewReadFilesWorkflow(recordId, workflowId)
	result = append(result, readStep)
	lastStepName := readStep.Name

	for _, cfg := range conf {
		task := NewProcessingStep(recordId, workflowId, lastStepName, &TemplateReference{
			Name:     cfg.Name,
			Template: cfg.Template,
		})
		result = append(result, task)
		lastStepName = task.Name
	}

	writeStep := NewWriteWorkflow(recordId, workflowId, lastStepName)
	result = append(result, writeStep)

	return result
}

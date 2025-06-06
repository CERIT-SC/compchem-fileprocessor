package argodtos

import (
	"fmt"
	"strings"

	"fi.muni.cz/invenio-file-processor/v2/config"
)

type WorkflowWrapper struct {
	Workflow Workflow `json:"workflow"`
}

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

func constructWorkflowName(workflowName string, recordId string, workflowId uint64) string {
	return fmt.Sprintf("%s-%s-%d", workflowName, recordId, workflowId)
}

func BuildWorkflow(
	conf config.WorkflowConfig,
	baseUrl string,
	workflowName string,
	workflowId uint64,
	recordId string,
	fileIds []string,
) *Workflow {
	tasks := constructLinearDag(conf.ProcessingTemplates, workflowName, recordId, workflowId)

	return newWorkflow(workflowName, recordId, baseUrl, workflowId, fileIds, tasks)
}

func newWorkflow(workflowName string,
	recordId string,
	baseUrl string,
	workflowId uint64,
	fileIds []string,
	processingTasks []*Task,
) *Workflow {
	fullName := constructWorkflowName(workflowName, recordId, workflowId)
	return &Workflow{
		ApiVersion: "argoproj.io/v1alpha1",
		Kind:       "Workflow",
		Metadata: Metadata{
			Name: fullName,
		},
		Spec: Spec{
			Entrypoint: fullName,
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
					Name: fullName,
					Dag: Dag{
						Tasks: processingTasks,
					},
				},
			},
		},
	}
}

func constructLinearDag(
	conf []config.ProcessingTemplate,
	worfklowName string,
	recordId string,
	workflowId uint64,
) []*Task {
	result := []*Task{}

	readStep := NewReadFilesWorkflow(recordId, workflowId)
	result = append(result, readStep)

	// each processing task executed after write task with its own write task
	for _, cfg := range conf {
		task := NewProcessingStep(recordId, workflowId, readStep.Name, &TemplateReference{
			Name:     cfg.Name,
			Template: cfg.Template,
		})
		writeTask := NewWriteWorkflow(
			recordId,
			workflowId,
			task.Name,
			cfg.Template,
			constructWorkflowName(worfklowName, recordId, workflowId),
		)
		result = append(result, task, writeTask)
	}

	return result
}

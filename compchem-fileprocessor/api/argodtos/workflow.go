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

func ConstructFullWorkflowName(workflowName string, recordId string, workflowId uint64) string {
	return fmt.Sprintf("%s-%s-%d", workflowName, recordId, workflowId)
}

func BuildWorkflow(
	conf config.WorkflowConfig,
	baseUrl string,
	workflowName string,
	workflowId uint64,
	secretKey string,
	recordId string,
	fileIds []string,
) *Workflow {
	tasks := constructLinearDag(
		conf.ProcessingTemplates,
		workflowName,
		recordId,
		workflowId,
		secretKey,
	)

	return newWorkflow(workflowName, recordId, baseUrl, workflowId, secretKey, fileIds, tasks)
}

func newWorkflow(workflowName string,
	recordId string,
	baseUrl string,
	workflowId uint64,
	secretKey string,
	fileIds []string,
	processingTasks []*Task,
) *Workflow {
	fullName := ConstructFullWorkflowName(workflowName, recordId, workflowId)
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
						Name:  "record-id",
						Value: recordId,
					},
					{
						Name:  "secret-key",
						Value: secretKey,
					},
					{
						Name:  "file-ids",
						Value: strings.Join(fileIds, " "),
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
	secretKey string,
) []*Task {
	result := []*Task{}
	deleteDependencies := []string{}

	readStep := newReadFilesWorkflow(recordId, workflowId)
	result = append(result, readStep)
	fullWorkflowName := ConstructFullWorkflowName(worfklowName, recordId, workflowId)

	// each processing task executed after write task with its own write task
	for _, cfg := range conf {
		task := newProcessingStep(recordId, workflowId, readStep.Name, &TemplateReference{
			Name:     cfg.Name,
			Template: cfg.Template,
		})
		writeTask := newWriteWorkflow(
			recordId,
			workflowId,
			task.Name,
			cfg.Template,
			fullWorkflowName,
		)
		result = append(result, task, writeTask)
		deleteDependencies = append(deleteDependencies, writeTask.Name)
	}

	revokeTokenStep := newDeleteWorkflow(
		recordId,
		workflowId,
		fullWorkflowName,
		secretKey,
		deleteDependencies,
	)
	result = append(result, revokeTokenStep)

	return result
}

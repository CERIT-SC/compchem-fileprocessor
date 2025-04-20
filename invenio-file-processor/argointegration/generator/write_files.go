package generator

import "fmt"

type writeWorkflow struct {
	ApiVersion string   `json:"apiVersion"`
	Kind       string   `json:"kind"`
	Metadata   Metadata `json:"metadata"`
	Spec       Spec     `json:"spec"`
}

type Metadata struct {
	GenerateName string `json:"generateName"`
}

type Spec struct {
	WorkflowTemplateRef WorkflowTemplateRef `json:"workflowTemplateRef"`
	Arguments           Arguments           `json:"arguments"`
}

type WorkflowTemplateRef struct {
	Name string `json:"name"`
}

type Arguments struct {
	Paramaters []KeyVal   `json:"paramaters"`
	Artifacts  []Artifact `json:"artifacts"`
}

type KeyVal struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Artifact struct {
	Name string `json:"name"`
	From string `json:"from"`
}

func WriteWorkflow(
	recordId string,
	workflowId string,
	baseUrl string,
	fileIds []string,
) *writeWorkflow {
	return &writeWorkflow{
		ApiVersion: "argoproj.io/v1alpha1",
		Kind:       "Workflow",
		Metadata: Metadata{
			GenerateName: fmt.Sprintf("write-files-%s-%s", recordId, workflowId),
		},
		Spec: Spec{
			WorkflowTemplateRef: WorkflowTemplateRef{
				Name: "write-files-template",
			},
			Arguments: Arguments{
				Paramaters: []KeyVal{
					{
						Name:  "base-url",
						Value: baseUrl,
					},
					{
						Name:  "record-id",
						Value: recordId,
					},
				},
				Artifacts: []Artifact{
					{
						Name: "files",
						From: "{{steps.read.outputs.Artifacts.downloaded-files}}",
					},
				},
			},
		},
	}
}

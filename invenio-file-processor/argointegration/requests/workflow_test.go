package requests

import (
	"encoding/json"
	"testing"

	"fi.muni.cz/invenio-file-processor/v2/config"
	"github.com/stretchr/testify/assert"
)

func TestBuildWorkflow_IntegrationTest(t *testing.T) {
	// Arrange
	recordId := "12345"
	workflowId := "2"
	workflowName := "read-count-write"
	baseUrl := "https://host-service.argo.svc.cluster.local:5000/api/experiments"
	
	processingTemplates := []config.ProcessingTemplate{
		{
			Name:     "count-words-template",
			Template: "count-words-",
		},
		{
			Name:     "count-words-advanced-template",
			Template: "count-words-advanced-",
		},
	}
	
	workflowConfig := config.WorkflowConfig{
		ProcessingTemplates: processingTemplates,
	}
	
	// Act
	workflow := BuildWorkflow(workflowConfig, baseUrl, workflowName, workflowId, recordId, []string{"test.txt", "test1.txt"})
	
	// Assert - In-memory representation
	
	// Check metadata
	assert.Equal(t, "argoproj.io/v1alpha1", workflow.ApiVersion)
	assert.Equal(t, "Workflow", workflow.Kind)
	assert.Equal(t, "read-count-write-12345-2", workflow.Metadata.Name)
	
	// Check spec
	assert.Equal(t, "read-count-write-12345-2", workflow.Spec.Entrypoint)
	
	// Check arguments
	assert.Equal(t, 3, len(workflow.Spec.Arguments.Parameters))
	assert.Equal(t, "base-url", workflow.Spec.Arguments.Parameters[0].Name)
	assert.Equal(t, baseUrl, workflow.Spec.Arguments.Parameters[0].Value)
	assert.Equal(t, "file-ids", workflow.Spec.Arguments.Parameters[1].Name)
	assert.Equal(t, "test.txt test1.txt", workflow.Spec.Arguments.Parameters[1].Value) // Empty string as no file IDs were provided
	assert.Equal(t, "record-id", workflow.Spec.Arguments.Parameters[2].Name)
	assert.Equal(t, recordId, workflow.Spec.Arguments.Parameters[2].Value)
	
	// Check templates
	assert.Equal(t, 1, len(workflow.Spec.Templates))
	assert.Equal(t, "read-count-write-12345-2", workflow.Spec.Templates[0].Name)
	
	// Check tasks
	tasks := workflow.Spec.Templates[0].Dag.Tasks
	assert.Equal(t, 4, len(tasks))
	
	// Check first task (read)
	assert.Equal(t, "read-files-12345-2", tasks[0].Name)
	assert.Equal(t, "[]", tasks[0].Dependencies)
	assert.Equal(t, "read-files-template", tasks[0].TemplateReference.Name)
	assert.Equal(t, "read-files", tasks[0].TemplateReference.Template)
	assert.Equal(t, 3, len(tasks[0].Arguments.Parameters))
	assert.Equal(t, 0, len(tasks[0].Arguments.Artifacts))
	
	// Check second task (first count)
	assert.Equal(t, "count-words-12345-2", tasks[1].Name)
	assert.Equal(t, "[read-files-12345-2]", tasks[1].Dependencies)
	assert.Equal(t, "count-words-template", tasks[1].TemplateReference.Name)
	assert.Equal(t, "count-words-", tasks[1].TemplateReference.Template)
	assert.Equal(t, 0, len(tasks[1].Arguments.Parameters))
	assert.Equal(t, 1, len(tasks[1].Arguments.Artifacts))
	assert.Equal(t, "input-files", tasks[1].Arguments.Artifacts[0].Name)
	assert.Equal(t, "{{tasks.read-files-12345-2.outputs.artifacts.output-files}}", tasks[1].Arguments.Artifacts[0].From)
	
	// Check third task (second count)
	assert.Equal(t, "count-words-advanced-12345-2", tasks[2].Name)
	assert.Equal(t, "[count-words-12345-2]", tasks[2].Dependencies)
	assert.Equal(t, "count-words-advanced-template", tasks[2].TemplateReference.Name)
	assert.Equal(t, "count-words-advanced-", tasks[2].TemplateReference.Template)
	assert.Equal(t, 0, len(tasks[2].Arguments.Parameters))
	assert.Equal(t, 1, len(tasks[2].Arguments.Artifacts))
	assert.Equal(t, "input-files", tasks[2].Arguments.Artifacts[0].Name)
	assert.Equal(t, "{{tasks.count-words-12345-2.outputs.artifacts.output-files}}", tasks[2].Arguments.Artifacts[0].From)
	
	// Check fourth task (write)
	assert.Equal(t, "write-files-12345-2", tasks[3].Name)
	assert.Equal(t, "[count-words-advanced-12345-2]", tasks[3].Dependencies)
	assert.Equal(t, "write-files-template", tasks[3].TemplateReference.Name)
	assert.Equal(t, "upload-files", tasks[3].TemplateReference.Template)
	assert.Equal(t, 2, len(tasks[3].Arguments.Parameters))
	assert.Equal(t, 1, len(tasks[3].Arguments.Artifacts))
	assert.Equal(t, "input-files", tasks[3].Arguments.Artifacts[0].Name)
	assert.Equal(t, "{{tasks.count-words-advanced-12345-2.outputs.artifacts.output-files}}", tasks[3].Arguments.Artifacts[0].From)
	
	// Assert - JSON serialization
	workflowJson, err := json.Marshal(workflow)
	assert.NoError(t, err)
	
	expectedJson := `{
		"apiVersion": "argoproj.io/v1alpha1",
		"kind": "Workflow",
		"metadata": {
			"name": "read-count-write-12345-2"
		},
		"spec": {
			"entrypoint": "read-count-write-12345-2",
			"arguments": {
				"parameters": [
					{
						"name": "base-url",
						"value": "https://host-service.argo.svc.cluster.local:5000/api/experiments"
					},
					{
						"name": "file-ids",
						"value": "test.txt test1.txt"
					},
					{
						"name": "record-id",
						"value": "12345"
					}
				]
			},
			"templates": [
				{
					"name": "read-count-write-12345-2",
					"dag": {
						"tasks": [
							{
								"name": "read-files-12345-2",
								"dependencies": "[]",
								"templateRef": {
									"name": "read-files-template",
									"template": "read-files"
								},
								"arguments": {
									"parameters": [
										{
											"name": "base-url",
											"value": "{{workflow.parameters.base-url}}"
										},
										{
											"name": "record-id",
											"value": "{{workflow.parameters.record-id}}"
										},
										{
											"name": "file-ids",
											"value": "{{workflow.parameters.file-ids}}"
										}
									],
									"artifacts": []
								}
							},
							{
								"name": "count-words-12345-2",
								"dependencies": "[read-files-12345-2]",
								"templateRef": {
									"name": "count-words-template",
									"template": "count-words-"
								},
								"arguments": {
									"parameters": [],
									"artifacts": [
										{
											"name": "input-files",
											"from": "{{tasks.read-files-12345-2.outputs.artifacts.output-files}}"
										}
									]
								}
							},
							{
								"name": "count-words-advanced-12345-2",
								"dependencies": "[count-words-12345-2]",
								"templateRef": {
									"name": "count-words-advanced-template",
									"template": "count-words-advanced-"
								},
								"arguments": {
									"parameters": [],
									"artifacts": [
										{
											"name": "input-files",
											"from": "{{tasks.count-words-12345-2.outputs.artifacts.output-files}}"
										}
									]
								}
							},
							{
								"name": "write-files-12345-2",
								"dependencies": "[count-words-advanced-12345-2]",
								"templateRef": {
									"name": "write-files-template",
									"template": "upload-files"
								},
								"arguments": {
									"parameters": [
										{
											"name": "base-url",
											"value": "{{workflow.parameters.base-url}}"
										},
										{
											"name": "record-id",
											"value": "{{workflow.parameters.record-id}}"
										}
									],
									"artifacts": [
										{
											"name": "input-files",
											"from": "{{tasks.count-words-advanced-12345-2.outputs.artifacts.output-files}}"
										}
									]
								}
							}
						]
					}
				}
			]
		}
	}`
	
	// Normalize both JSON strings for comparison
	var expected, actual any 
	err = json.Unmarshal([]byte(expectedJson), &expected)
	assert.NoError(t, err)
	
	err = json.Unmarshal(workflowJson, &actual)
	assert.NoError(t, err)
	
	// Re-marshal both to normalized format
	expectedNormalized, err := json.Marshal(expected)
	assert.NoError(t, err)
	
	actualNormalized, err := json.Marshal(actual)
	assert.NoError(t, err)
	
	// Compare the normalized JSON strings
	assert.Equal(t, string(expectedNormalized), string(actualNormalized))
}

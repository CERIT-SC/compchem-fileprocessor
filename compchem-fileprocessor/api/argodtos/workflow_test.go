package argodtos

import (
	"encoding/json"
	"testing"

	"fi.muni.cz/invenio-file-processor/v2/config"
	"github.com/stretchr/testify/assert"
)

func TestBuildWorkflow_IntegrationTest(t *testing.T) {
	// Arrange
	recordId := "12345"
	workflowId := uint64(2)
	workflowName := "read-count-write-12345-2"
	baseUrl := "https://host-service.argo.svc.cluster.local:5000/api/experiments"

	processingTemplates := []config.ProcessingTemplate{
		{
			Name:     "count-words-template",
			Template: "count-words",
		},
		{
			Name:     "count-words-advanced-template",
			Template: "count-words-advanced",
		},
	}

	workflowConfig := config.WorkflowConfig{
		ProcessingTemplates: processingTemplates,
	}

	// Act
	workflow := BuildWorkflow(
		workflowConfig,
		baseUrl,
		workflowName,
		workflowId,
		recordId,
		[]string{"test.txt", "test1.txt"},
	)

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
								"dependencies": [],
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
								"dependencies": ["read-files-12345-2"],
								"templateRef": {
									"name": "count-words-template",
									"template": "count-words"
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
								"name": "write-files-count-words-12345-2",
								"dependencies": ["count-words-12345-2"],
								"templateRef": {
									"name": "write-files-template",
									"template": "write-files"
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
											"name": "workflow-name",
											"value": "read-count-write-12345-2"
										},
										{
											"name": "task-discriminator",
											"value": "count-words"
										}
									],
									"artifacts": [
										{
											"name": "input-files",
											"from": "{{tasks.count-words-12345-2.outputs.artifacts.output-files}}"
										}
									]
								}
							},
							{
								"name": "count-words-advanced-12345-2",
								"dependencies": ["read-files-12345-2"],
								"templateRef": {
									"name": "count-words-advanced-template",
									"template": "count-words-advanced"
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
								"name": "write-files-count-words-advanced-12345-2",
								"dependencies": ["count-words-advanced-12345-2"],
								"templateRef": {
									"name": "write-files-template",
									"template": "write-files"
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
											"name": "workflow-name",
											"value": "read-count-write-12345-2"
										},
										{
											"name": "task-discriminator",
											"value": "count-words-advanced"
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

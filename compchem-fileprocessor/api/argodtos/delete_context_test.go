package argodtos

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteWorkflow_AllArgumentsSupplied_ProperlyFormedTask(t *testing.T) {
	// Arrange
	recordId := "12345"
	workflowId := uint64(2)
	workflowFullName := "read-count-write-12345-2"
	previousTasks := []string{"write-files-task1", "write-files-task2"}

	expectedName := fmt.Sprintf(deleteContextTemplate, recordId, workflowId)
	expectedDependencies := previousTasks

	// Act
	task := newDeleteWorkflow(recordId, workflowId, workflowFullName, previousTasks)

	// Assert
	assert.Equal(t, expectedName, task.Name)
	assert.Equal(t, expectedDependencies, task.Dependencies)
	assert.Equal(t, "delete-context-template", task.TemplateReference.Name)
	assert.Equal(t, "delete-context", task.TemplateReference.Template)

	// Verify artifacts is empty
	assert.Equal(t, 0, len(task.Arguments.Artifacts))

	// Verify parameters
	assert.Equal(t, 3, len(task.Arguments.Parameters))

	// Check base-url parameter
	assert.Equal(t, "base-url", task.Arguments.Parameters[0].Name)
	assert.Equal(t, "{{workflow.parameters.base-url}}", task.Arguments.Parameters[0].Value)

	// Check workflow-name parameter
	assert.Equal(t, "workflow-name", task.Arguments.Parameters[1].Name)
	assert.Equal(t, workflowFullName, task.Arguments.Parameters[1].Value)

	// Check secret-key parameter
	assert.Equal(t, "secret-key", task.Arguments.Parameters[2].Name)
	assert.Equal(t, "{{workflow.parameters.secret-key}}", task.Arguments.Parameters[2].Value)
}

func TestDeleteWorkflow_AllArgumentsSupplied_ProperlyFormedJson(t *testing.T) {
	// Arrange
	recordId := "12345"
	workflowId := uint64(2)
	workflowFullName := "read-count-write-12345-2"
	previousTasks := []string{"write-files-task1", "write-files-task2"}

	task := newDeleteWorkflow(recordId, workflowId, workflowFullName, previousTasks)

	// Act
	taskJson, err := json.Marshal(task)

	// Assert
	assert.NoError(t, err)

	// Define expected JSON
	expectedJson := `{
		"name": "delete-context-12345-2",
		"dependencies": ["write-files-task1", "write-files-task2"],
		"templateRef": {
			"name": "delete-context-template",
			"template": "delete-context"
		},
		"arguments": {
			"parameters": [
				{
					"name": "base-url",
					"value": "{{workflow.parameters.base-url}}"
				},
				{
					"name": "workflow-name",
					"value": "read-count-write-12345-2"
				},
				{
					"name": "secret-key",
					"value": "{{workflow.parameters.secret-key}}"
				}
			],
			"artifacts": []
		}
	}`

	// Normalize both JSON strings for comparison
	var expected, actual any
	err = json.Unmarshal([]byte(expectedJson), &expected)
	assert.NoError(t, err)

	err = json.Unmarshal(taskJson, &actual)
	assert.NoError(t, err)

	// Re-marshal both to normalized format
	expectedNormalized, err := json.Marshal(expected)
	assert.NoError(t, err)

	actualNormalized, err := json.Marshal(actual)
	assert.NoError(t, err)

	// Compare the normalized JSON strings
	assert.Equal(t, string(expectedNormalized), string(actualNormalized))
}

func TestDeleteWorkflow_EmptyPreviousTasks_EmptyDependencies(t *testing.T) {
	// Arrange
	recordId := "54321"
	workflowId := uint64(5)
	workflowFullName := "test-workflow-54321-5"
	previousTasks := []string{}

	// Act
	task := newDeleteWorkflow(recordId, workflowId, workflowFullName, previousTasks)

	// Assert
	assert.Equal(t, fmt.Sprintf(deleteContextTemplate, recordId, workflowId), task.Name)
	assert.Equal(t, []string{}, task.Dependencies)
	assert.Equal(t, workflowFullName, task.Arguments.Parameters[1].Value)
}

func TestDeleteWorkflow_SinglePreviousTask_SingleDependency(t *testing.T) {
	// Arrange
	recordId := "99999"
	workflowId := uint64(1)
	workflowFullName := "single-task-99999-1"
	previousTasks := []string{"write-files-single-task"}

	// Act
	task := newDeleteWorkflow(recordId, workflowId, workflowFullName, previousTasks)

	// Assert
	assert.Equal(t, fmt.Sprintf(deleteContextTemplate, recordId, workflowId), task.Name)
	assert.Equal(t, previousTasks, task.Dependencies)
	assert.Equal(t, 1, len(task.Dependencies))
	assert.Equal(t, "write-files-single-task", task.Dependencies[0])
}

func TestDeleteTokenTemplate_ConstantValue(t *testing.T) {
	// Test the constant template format
	assert.Equal(t, "delete-context-%s-%d", deleteContextTemplate)
}

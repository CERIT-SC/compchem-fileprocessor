package argodtos

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// templated by me generated by claude AI
func TestProcessingStep_AllArgumentsSupplied_ProperlyFormedTask(t *testing.T) {
	// Arrange
	recordId := "12345"
	workflowId := uint64(2)
	previousTask := "read-files-12345-2"
	templateRef := &TemplateReference{
		Name:     "count-words-template",
		Template: "count-words",
	}

	expectedName := fmt.Sprintf("count-words-%s-%d", recordId, workflowId)
	expectedDependencies := []string{previousTask}

	// Act
	task := NewProcessingStep(recordId, workflowId, previousTask, templateRef)

	// Assert
	assert.Equal(t, expectedName, task.Name)
	assert.Equal(t, expectedDependencies, task.Dependencies)
	assert.Equal(t, templateRef.Name, task.TemplateReference.Name)
	assert.Equal(t, templateRef.Template, task.TemplateReference.Template)

	// Verify parameters is empty
	assert.Equal(t, 0, len(task.Arguments.Parameters))

	// Verify artifacts
	assert.Equal(t, 1, len(task.Arguments.Artifacts))
	assert.Equal(t, "input-files", task.Arguments.Artifacts[0].Name)
	assert.Equal(
		t,
		fmt.Sprintf("{{tasks.%s.outputs.artifacts.output-files}}", previousTask),
		task.Arguments.Artifacts[0].From,
	)
}

func TestProcessingStep_AllArgumentsSupplied_ProperlyFormedJson(t *testing.T) {
	// Arrange
	recordId := "12345"
	workflowId := uint64(2)
	previousTask := "read-files-12345-2"
	templateRef := &TemplateReference{
		Name:     "count-words-template",
		Template: "count-words",
	}

	task := NewProcessingStep(recordId, workflowId, previousTask, templateRef)

	// Act
	taskJson, err := json.Marshal(task)

	// Assert
	assert.NoError(t, err)

	// Define expected JSON
	expectedJson := `{
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
	}`

	// Normalize both JSON strings for comparison (remove whitespace differences)
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

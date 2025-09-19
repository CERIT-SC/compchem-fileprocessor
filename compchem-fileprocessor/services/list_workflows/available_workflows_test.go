package list_workflows

import (
	"testing"

	"fi.muni.cz/invenio-file-processor/v2/api/availabledtos"
	"fi.muni.cz/invenio-file-processor/v2/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestConvertRequest_ThreeMimeTypes_ThreeMapKeys(t *testing.T) {
	TEXT_PLAIN := "text/plain"
	JSON := "application/json"
	IMAGE := "image/jpeg"

	request := availabledtos.AvailableWorkflowsRequest{
		Files: []availabledtos.KeyAndType{
			{
				FileKey:  "test1.txt",
				Mimetype: TEXT_PLAIN,
			},
			{
				FileKey:  "test2.txt",
				Mimetype: TEXT_PLAIN,
			},
			{
				FileKey:  "test.jpeg",
				Mimetype: IMAGE,
			},
			{
				FileKey:  "request.json",
				Mimetype: JSON,
			},
		},
	}

	mimetypeMap := convertRequestToMap(&request)

	assert.Equal(t, len(mimetypeMap), len(mimetypeMap))
	assert.Equal(t, len(mimetypeMap[TEXT_PLAIN]), len(mimetypeMap[TEXT_PLAIN]))
	assert.Equal(t, len(mimetypeMap[IMAGE]), len(mimetypeMap[IMAGE]))
	assert.Equal(t, len(mimetypeMap[JSON]), len(mimetypeMap[JSON]))

	assert.Contains(t, mimetypeMap[TEXT_PLAIN], "test1.txt")
	assert.Contains(t, mimetypeMap[TEXT_PLAIN], "test2.txt")
	assert.Contains(t, mimetypeMap[IMAGE], "test.jpeg")
	assert.Contains(t, mimetypeMap[JSON], "request.json")
}

func TestCreateResponse_TwoConfigs_CorrectResponse(t *testing.T) {
	configs := []config.WorkflowConfig{
		{
			Name:     "count-words",
			Filetype: "text/plain",
			ProcessingTemplates: []config.ProcessingTemplate{
				{
					Name:     "count-words",
					Template: "count-words-template",
				},
			},
		},
		{
			Name:     "downsizeto480p",
			Filetype: "image/jpeg",
			ProcessingTemplates: []config.ProcessingTemplate{
				{
					Name:     "downsizepng",
					Template: "downsizepng-template",
				},
				{
					Name:     "downsizejpeg",
					Template: "downsizejpeg-template",
				},
			},
		},
	}

	mimetypeMap := map[string][]string{
		"text/plain":      {"mytextfile.txt", "mysecondtext.txt"},
		"application/pdf": {"mypdffile.txt"},
	}

	result := convertMapToAvailableWorkflows(mimetypeMap, configs)

	// only for txt
	assert.Len(t, result.Workflows, 1)
	assert.Len(t, result.Workflows[0].Files, 2)
	assert.Contains(t, result.Workflows[0].Files, "mytextfile.txt")
	assert.Contains(t, result.Workflows[0].Files, "mysecondtext.txt")
	assert.Equal(t, result.Workflows[0].Mimetype, "text/plain")
}

func TestAvailableWorkflows_TwoConfigs_ReturnsCorrectResponse(t *testing.T) {
	TEXT_PLAIN := "text/plain"
	JSON := "application/json"
	IMAGE := "image/jpeg"

	request := availabledtos.AvailableWorkflowsRequest{
		Files: []availabledtos.KeyAndType{
			{
				FileKey:  "test1.txt",
				Mimetype: TEXT_PLAIN,
			},
			{
				FileKey:  "test2.txt",
				Mimetype: TEXT_PLAIN,
			},
			{
				FileKey:  "test.jpeg",
				Mimetype: IMAGE,
			},
			{
				FileKey:  "request.json",
				Mimetype: JSON,
			},
		},
	}

	configs := []config.WorkflowConfig{
		{
			Name:     "count-words",
			Filetype: "text/plain",
			ProcessingTemplates: []config.ProcessingTemplate{
				{
					Name:     "count-words",
					Template: "count-words-template",
				},
			},
		},
		{
			Name:     "downsizeto480p",
			Filetype: "image/jpeg",
			ProcessingTemplates: []config.ProcessingTemplate{
				{
					Name:     "downsizepng",
					Template: "downsizepng-template",
				},
				{
					Name:     "downsizejpeg",
					Template: "downsizejpeg-template",
				},
			},
		},
	}

	response := AvailableWorkflows(zap.NewNop(), &request, configs)

	assert.Len(t, response.Workflows, 2)

	for _, availableWorkflow := range response.Workflows {
		if availableWorkflow.Mimetype == "text/plain" {
			assert.Len(t, availableWorkflow.Files, 2)
			assert.Contains(t, availableWorkflow.Files, "test1.txt")
			assert.Contains(t, availableWorkflow.Files, "test2.txt")
		} else if availableWorkflow.Mimetype == "image/jpeg" {
			assert.Len(t, availableWorkflow.Files, 1)
			assert.Equal(t, availableWorkflow.Files[0], "test.jpeg")
		}
	}
}

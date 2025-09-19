package start_workflow_route

import (
	"net/http/httptest"
	"strings"
	"testing"

	"fi.muni.cz/invenio-file-processor/v2/routes/common"
	startworkflow_service "fi.muni.cz/invenio-file-processor/v2/services/start_workflow"
	"github.com/stretchr/testify/assert"
)

func TestValidateBody_MissingBody_ReturnsError(t *testing.T) {
	reader := strings.NewReader(`
  {
    "key": "test",
    "mimetype": "test"
  }
  `)

	recorder := httptest.NewRecorder()

	request := httptest.NewRequest("POST", "https://localhost:8080", reader)

	reqBody, err := common.GetValidRequestBody(recorder, request, validateStartBody)
	assert.Nil(t, reqBody, "body should be nil")
	assert.Error(t, err, "expected error returned")
}

func TestValidateBody_OkBody_ReturnsCorrectBody(t *testing.T) {
	expected := startRequestBody{
		Files: []startworkflow_service.File{
			{
				FileName: "test",
				Mimetype: "test",
			},
		},
		Name:     "count-words",
		RecordId: "ejw6-7fpy",
	}

	reader := strings.NewReader(`
  {
    "recordId": "ejw6-7fpy",
    "name": "count-words",
    "files": [
      {
        "key": "test",
        "mimetype": "test"
      }
    ]
  }
  `)

	recorder := httptest.NewRecorder()

	request := httptest.NewRequest("POST", "https://localhost:8080", reader)

	reqBody, err := common.GetValidRequestBody(recorder, request, validateStartBody)
	assert.Nil(t, err, "expected error returned")
	assert.Equal(t, expected, *reqBody, "expected same body as in test")
}

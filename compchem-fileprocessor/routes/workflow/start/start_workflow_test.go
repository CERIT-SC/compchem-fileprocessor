package start_workflow_route

import (
	"net/http/httptest"
	"strings"
	"testing"

	"fi.muni.cz/invenio-file-processor/v2/routes/common"
	"fi.muni.cz/invenio-file-processor/v2/service"
	"github.com/stretchr/testify/assert"
)

func TestValidateBody_MissingBody_ReturnsError(t *testing.T) {
	reader := strings.NewReader(`
  {
    "fileName": "test",
    "mimetype": "test"
  }
  `)

	recorder := httptest.NewRecorder()

	request := httptest.NewRequest("POST", "https://localhost:8080", reader)

	reqBody, err := common.GetValidRequestBody(recorder, request, validateBody)
	assert.Nil(t, reqBody, "body should be nil")
	assert.Error(t, err, "expected error returned")
}

func TestValidateBody_OkBody_ReturnsCorrectBody(t *testing.T) {
	expected := requestBody{
		Files: []service.File{
			{
				FileName: "test",
				Mimetype: "test",
			},
		},
		RecordId: "ejw6-7fpy",
	}

	reader := strings.NewReader(`
  {
    "recordId": "ejw6-7fpy",
    "files": [
      {
        "fileName": "test",
        "mimetype": "test"
      }
    ]
  }
  `)

	recorder := httptest.NewRecorder()

	request := httptest.NewRequest("POST", "https://localhost:8080", reader)

	reqBody, err := common.GetValidRequestBody(recorder, request, validateBody)
	assert.Nil(t, err, "expected error returned")
	assert.Equal(t, expected, *reqBody, "expected same body as in test")
}

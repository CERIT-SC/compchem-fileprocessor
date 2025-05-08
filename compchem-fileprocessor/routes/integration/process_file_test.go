package integration

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateBody_MissingBody_ReturnsError(t *testing.T) {
	reader := strings.NewReader(`
  {
    "fileName": "test",
    "fileType": "test"
  }
  `)

	recorder := httptest.NewRecorder()

	request := httptest.NewRequest("POST", "https://localhost:8080", reader)

	reqBody, err := getRequestBody(recorder, request)
	assert.Nil(t, reqBody, "body should be nil")
	assert.Error(t, err, "expected error returned")
}

func TestValidateBody_OkBody_ReturnsCorrectBody(t *testing.T) {
	expected := requestBody{
		FileName: "test",
		FileType: "test",
		RecordId: "ejw6-7fpy",
	}

	reader := strings.NewReader(`
  {
    "fileName": "test",
    "recordId": "ejw6-7fpy",
    "fileType": "test"
  }
  `)

	recorder := httptest.NewRecorder()

	request := httptest.NewRequest("POST", "https://localhost:8080", reader)

	reqBody, err := getRequestBody(recorder, request)
	assert.Nil(t, err, "expected error returned")
	assert.Equal(t, expected, *reqBody, "expected same body as in test")
}

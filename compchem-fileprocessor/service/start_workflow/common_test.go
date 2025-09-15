package startworkflow_service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetArgoUrl_ArgsProvided_UrlCorrectlyFormed(t *testing.T) {
	baseUrl := "https://argo-service.kubernetes.local"
	namespace := "argo"

	result := buildWorkflowUrl(namespace, baseUrl, "submit")
	assert.Equal(
		t,
		"https://argo-service.kubernetes.local/api/v1/workflows/argo/submit",
		result,
		"urls should match",
	)
}

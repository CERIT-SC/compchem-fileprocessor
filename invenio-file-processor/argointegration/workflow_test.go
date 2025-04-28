package argointegration

import (
	"testing"

	"fi.muni.cz/invenio-file-processor/v2/config"
	"github.com/stretchr/testify/assert"
)

func TestFindWorkflowConfig_MatchingConfigExists_ConfigFound(t *testing.T) {
	configs := []*config.WorkflowConfig{
		{
			Filetype: "txt",
		},
	}

	conf, err := findWorkflowConfig(configs, "txt")
	assert.NoError(t, err, "error should be nil because config exists")
	assert.Equal(t, conf, configs[0], "returned should be the same object as in setup")
}

func TestFindWorkflowConfig_NoConfig_ErorrReturned(t *testing.T) {
	configs := []*config.WorkflowConfig{
		{
			Filetype: "png",
		},
	}

	conf, err := findWorkflowConfig(configs, "txt")
	assert.Nil(t, conf, "config should be null")
	assert.Error(t, err, "error should have been returned")
}

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

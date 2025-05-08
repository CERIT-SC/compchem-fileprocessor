package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestReadConfig_NoFile_ErrPropagated(t *testing.T) {
	logger := zaptest.NewLogger(t)

	_, err := readConfig(logger, "non-existent-config.yaml")
	if err == nil {
		t.Fatal("Expected error when reading non-existent config file, got nil")
	}
}

func TestReadConfig_MockInMemoryFile_BytesReturned(t *testing.T) {
	logger := zaptest.NewLogger(t)

	tmpDir := t.TempDir()
	mockPath := filepath.Join(tmpDir, "mock.yaml")
	content := []byte("host: localhost\nport: 1234")

	err := os.WriteFile(mockPath, content, 0644)
	if err != nil {
		t.Fatalf("Failed to write mock config file: %v", err)
	}

	data, err := readConfig(logger, mockPath)
	if err != nil {
		t.Fatalf("Unexpected error reading mock config file: %v", err)
	}

	if string(data) != string(content) {
		t.Errorf("Read bytes do not match. Got: %s", data)
	}
}

func TestResolveConfig_OkConfig_ConfigResolutionOk(t *testing.T) {
	logger := zaptest.NewLogger(t)

	yamlBytes := []byte(`
server:
  host: example.com
  port: 9999
`)

	cfg, err := resolveConfig(logger, yamlBytes)
	if err != nil {
		t.Fatalf("Unexpected error resolving config: %v", err)
	}

	if cfg.Server.Host != "example.com" || cfg.Server.Port != 9999 {
		t.Errorf("Config values incorrect: got %+v", cfg)
	}
}

func TestResolveConfig_ConfigMalformed_ConfigResolutionError(t *testing.T) {
	logger := zaptest.NewLogger(t)

	yamlBytes := []byte(`
server:
  host: localhost
  port: not-a-number
`)

	_, err := resolveConfig(logger, yamlBytes)
	if err == nil {
		t.Fatal("Expected error when resolving malformed config, got nil")
	}
}

func TestValidateConfig_ConfigOk_ConfigValidationOk(t *testing.T) {
	logger := zaptest.NewLogger(t)

	cfg := &Config{
		Server: Server{
			Host: "127.0.0.1",
			Port: 8080,
		},
		ArgoApi: ArgoApi{
			Url:       "https://localhost:2746",
			Namespace: "test",
		},
		CompchemApi: CompchemApi{
			Url: "https://localhost:5000",
		},
	}

	result, errs := validateConfig(logger, cfg)
	if len(errs) > 0 {
		t.Errorf("Expected no validation errors, got: %v", errs)
	}

	if result.Server.Host != "127.0.0.1" || result.Server.Port != 8080 {
		t.Errorf("Unexpected config values after validation: %+v", result)
	}
}

func TestValidateConfig_ConfigNotOk_ConfigValidationNotOk(t *testing.T) {
	logger := zaptest.NewLogger(t)

	cfg := &Config{
		Server: Server{
			Host: "127.0.0.1",
			Port: 8080,
		},
		Workflows: []WorkflowConfig{
			{
				Name: "test",
			},
		},
	}

	result, errs := validateConfig(logger, cfg)
	if len(errs) != 4 {
		t.Errorf("Expected 4 validation errors, got: %v", errs)
	}

	if result.Server.Host != "127.0.0.1" || result.Server.Port != 8080 {
		t.Errorf("Unexpected config values after validation: %+v", result)
	}

	if len(result.Workflows) != 1 {
		t.Errorf("Expected one workflow defined")
	}
}

func TestLoadConfig_ConfigOk_ConfigLoadedCorrectly(t *testing.T) {
	tmpDir := t.TempDir()
	mockPath := filepath.Join(tmpDir, "server-config.yaml")
	content := []byte(`
server:
  host: localhost
  port: 8062

context-path: "/api"

argo-workflows:
  url: https://localhost:2746
  namespace: "argo"

compchem:
  url: https://localhost:5000

workflows:
  - name: count-words
    filetype: txt
    processing-templates:
      - name: count-words-template
        template: count-words
  `)

	err := os.WriteFile(mockPath, content, 0644)
	if err != nil {
		t.Fatalf("Failed to write mock config file: %v", err)
	}

	config, err := LoadConfig(zap.NewNop(), tmpDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check Server configuration
	assert.Equal(t, "localhost", config.Server.Host, "Server host should match")
	assert.Equal(t, 8062, config.Server.Port, "Server port should match")

	// Check ApiContext
	assert.Equal(t, "/api", config.ApiContext, "API context path should match")

	// Check ArgoApi
	assert.Equal(t, "https://localhost:2746", config.ArgoApi.Url, "Argo API URL should match")
	assert.Equal(t, "argo", config.ArgoApi.Namespace, "Argo namespace should match")

	// Check CompchemApi
	assert.Equal(
		t,
		"https://localhost:5000",
		config.CompchemApi.Url,
		"Compchem API URL should match",
	)

	// Check Workflows
	assert.Equal(t, 1, len(config.Workflows), "Should have 1 workflow configured")
	assert.Equal(t, "count-words", config.Workflows[0].Name, "Workflow name should match")
	assert.Equal(t, "txt", config.Workflows[0].Filetype, "Workflow filetype should match")

	assert.Equal(
		t,
		1,
		len(config.Workflows[0].ProcessingTemplates),
		"Should have 1 processing template",
	)
	assert.Equal(
		t,
		"count-words-template",
		config.Workflows[0].ProcessingTemplates[0].Name,
		"Processing template name should match",
	)
	assert.Equal(
		t,
		"count-words",
		config.Workflows[0].ProcessingTemplates[0].Template,
		"Processing template value should match",
	)
}

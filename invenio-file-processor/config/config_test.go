package config

import (
	"os"
	"path/filepath"
	"testing"

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
	// For now we don't care — test stub only
}

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server      Server      `yaml:"server"`
	ApiContext  string      `yaml:"context-path"`
	CompchemApi CompchemApi `yaml:"compchem"`
	ArgoApi     ArgoApi     `yaml:"argo-workflows"`
  Workflows []WorkflowConfig `yaml:"workflows"`
}

type Server struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type CompchemApi struct {
	Url string `yaml:"url"`
}

type ArgoApi struct {
	Url       string `yaml:"url"`
	Namespace string `yaml:"namespace"`
}

type WorkflowConfig struct {
  Name string `yaml:"name"`
  Filetype string `yaml:"filetype"`
  ProcessingTemplates []ProcessingTemplate `yaml:"processing-template"`
}

type ProcessingTemplate struct {
  Name string `yaml:"name"`
  Template string `yaml:"template"`
}

func LoadConfig(logger *zap.Logger, executablePath string) (*Config, error) {
	DEFAULT_CONFIG_NAME := "server-config.yaml"
	configPath := filepath.Join(executablePath, DEFAULT_CONFIG_NAME)

	logger.Info("Loading config", zap.String("config_path", configPath))

	configBytes, err := readConfig(logger, configPath)
	if err != nil {
		return nil, err
	}

	config, err := resolveConfig(logger, configBytes)
	if err != nil {
		return nil, err
	}

	config, validationErrors := validateConfig(logger, config)
	if len(validationErrors) > 0 {
		logValidationErrors(logger, validationErrors)
		return nil, fmt.Errorf("config validation failed with %d error(s)", len(validationErrors))
	}

	return config, nil
}

func logValidationErrors(logger *zap.Logger, errors map[string]string) {
	for field, msg := range errors {
		logger.Error(
			"Config validation error",
			zap.String("field", field),
			zap.String("error", msg),
		)
	}
}

func readConfig(logger *zap.Logger, path string) ([]byte, error) {
	logger.Info("Reading config on path", zap.String("path", path))
	data, err := os.ReadFile(path)
	if err != nil {
		logger.Error("Read config file error", zap.Error(err))
		return nil, err
	}

	return data, nil
}

func resolveConfig(logger *zap.Logger, configBytes []byte) (*Config, error) {
	config := &Config{}

	err := yaml.Unmarshal(configBytes, config)
	if err != nil {
		logger.Error("rror during config resolution", zap.Error(err))
		return nil, err
	}

	return config, nil
}

func validateConfig(logger *zap.Logger, cfg *Config) (*Config, map[string]string) {
	errors := make(map[string]string)

	DEFAULT_HOST := "localhost"
	DEFAULT_PORT := 8079

	if cfg.Server.Host == "" {
		logger.Warn("Missing host, defaulting to localhost")
		cfg.Server.Host = DEFAULT_HOST
	}

	if cfg.Server.Port == 0 {
		logger.Warn("Missing port, defaulting to 8079")
		cfg.Server.Port = DEFAULT_PORT
	}

  if len(cfg.Workflows) > 0 {
    validateWorkflows(cfg.Workflows, errors)
  }

	return cfg, errors
}

func validateWorkflows(workflows []WorkflowConfig, errors map[string]string) {
  errorTemplate := "%s-%d"

  for index, workflow := range workflows {
    if workflow.Name == "" {
      errors[fmt.Sprintf(errorTemplate, "name", index)] = "missing name" 
    }
    if workflow.Filetype == "" {
      errors[fmt.Sprintf(errorTemplate, "filetype", index)] = "missing filetype" 
    }
    if len(workflow.ProcessingTemplates) > 0 {
      validateProcessingTemplates(workflow.ProcessingTemplates, index, errors)
    }
  }
}

func validateProcessingTemplates(templates []ProcessingTemplate, index int, errors map[string]string) {
  errorTemplate := "template-" + strconv.Itoa(index) + "-%d"

  for inx, template := range templates{
    if template.Name == "" {
      errors[fmt.Sprintf(errorTemplate, inx)] = "missing template name"
    }
    if template.Template == "" {
      errors[fmt.Sprintf(errorTemplate, inx)] = "missing template"
    } 
  }
}

package mcp

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"
)

// Config represents the configuration for an MCP server
type Config struct {
	// Name of the MCP server
	Name        string            `json:"name" yaml:"name"`
	Command     string            `json:"command" yaml:"command"`
	Args        []string          `json:"args,omitempty" yaml:"args,omitempty"`
	Environment map[string]string `json:"environment,omitempty" yaml:"environment,omitempty"`
}

// OllamaConfig represents the configuration for Ollama
type OllamaConfig struct {
	URL   string `json:"url" yaml:"url"`
	Model string `json:"model" yaml:"model"`
}

// ConfigFile represents the structure of the MCP configuration file
type ConfigFile struct {
	Servers []Config     `yaml:"servers"`
	Ollama  OllamaConfig `yaml:"ollama"`
}

// LoadConfigFromFile loads MCP server configurations from a YAML file
func LoadConfigFromFile(filePath string) ([]Config, error) {
	// Read the YAML file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", filePath, err)
	}

	// Parse the YAML
	var configFile ConfigFile
	if err := yaml.Unmarshal(data, &configFile); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	// Validate and process each server config
	for i, config := range configFile.Servers {
		if config.Name == "" {
			return nil, fmt.Errorf("server at index %d has empty name", i)
		}
		if config.Command == "" {
			return nil, fmt.Errorf("server %s has empty command", config.Name)
		}
	}

	return configFile.Servers, nil
}

// LoadConfigWithOllamaFromFile loads both MCP server and Ollama configurations from a YAML file
func LoadConfigWithOllamaFromFile(filePath string) ([]Config, OllamaConfig, error) {
	// Read the YAML file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, OllamaConfig{}, fmt.Errorf("failed to read config file %s: %w", filePath, err)
	}

	// Parse the YAML
	var configFile ConfigFile
	if err := yaml.Unmarshal(data, &configFile); err != nil {
		return nil, OllamaConfig{}, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	// Validate and process each server config
	for i, config := range configFile.Servers {
		if config.Name == "" {
			return nil, OllamaConfig{}, fmt.Errorf("server at index %d has empty name", i)
		}
		if config.Command == "" {
			return nil, OllamaConfig{}, fmt.Errorf("server %s has empty command", config.Name)
		}
	}

	// Set default values for Ollama if not provided
	if configFile.Ollama.URL == "" {
		configFile.Ollama.URL = "http://localhost:11434"
	}
	if configFile.Ollama.Model == "" {
		configFile.Ollama.Model = "llama3.2"
	}

	return configFile.Servers, configFile.Ollama, nil
}

// LoadConfigFromDefaultPath loads configuration from default paths
func LoadConfigFromDefaultPath() ([]Config, error) {
	// Try common configuration paths
	possiblePaths := []string{
		"mcp.yaml",
		"mcp.yml",
		"config/mcp.yaml",
		"config/mcp.yml",
	}

	// Try user home directory
	if homeDir, err := os.UserHomeDir(); err == nil {
		possiblePaths = append(possiblePaths,
			filepath.Join(homeDir, ".mcp.yaml"),
			filepath.Join(homeDir, ".mcp.yml"),
			filepath.Join(homeDir, ".config", "mcp.yaml"),
			filepath.Join(homeDir, ".config", "mcp.yml"),
		)
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return LoadConfigFromFile(path)
		}
	}

	return nil, fmt.Errorf("no MCP configuration file found in default paths")
}

// applyEnvironment applies environment variables to the configuration
func (c *Config) applyEnvironment() {
	if c.Environment == nil {
		return
	}

	// Set environment variables for this server
	for key, value := range c.Environment {
		expandedValue := expandEnvironmentVariables(value)
		os.Setenv(key, expandedValue)
	}
}

// CreateCommand creates an exec.Cmd with the configuration
func (c *Config) CreateCommand(ctx context.Context) *exec.Cmd {
	// Apply environment variables first
	c.applyEnvironment()

	// Expand environment variables in command and args
	expandedCommand := expandEnvironmentVariables(c.Command)
	expandedArgs := make([]string, len(c.Args))
	for i, arg := range c.Args {
		expandedArgs[i] = expandEnvironmentVariables(arg)
	}

	// Create the command
	cmd := exec.CommandContext(ctx, expandedCommand, expandedArgs...)

	// Set environment variables for the command
	if c.Environment != nil {
		env := os.Environ()
		for key, value := range c.Environment {
			expandedValue := expandEnvironmentVariables(value)
			env = append(env, fmt.Sprintf("%s=%s", key, expandedValue))
		}
		cmd.Env = env
	}

	return cmd
}

// expandEnvironmentVariables expands environment variables in the format ${VAR_NAME} or $VAR_NAME
func expandEnvironmentVariables(value string) string {
	// Handle ${VAR_NAME} format
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	expanded := re.ReplaceAllStringFunc(value, func(match string) string {
		varName := match[2 : len(match)-1] // Remove ${ and }
		return os.Getenv(varName)
	})

	// Handle $VAR_NAME format (simple cases)
	re2 := regexp.MustCompile(`\$([A-Za-z_][A-Za-z0-9_]*)`)
	expanded = re2.ReplaceAllStringFunc(expanded, func(match string) string {
		varName := match[1:] // Remove $
		return os.Getenv(varName)
	})

	return expanded
}

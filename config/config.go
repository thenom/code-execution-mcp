package config

import (
	"encoding/json"
	"os"
)

type MCPServerConfig struct {
	Name      string `json:"name"`
	URL       string `json:"url"`
	Timeout   int    `json:"timeout"`
	ExposeAll bool   `json:"expose_all"`
}

type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type ExecutionConfig struct {
	DefaultTimeout int    `json:"default_timeout"`
	MaxMemoryMB    int    `json:"max_memory_mb"`
	WorkspacePath  string `json:"workspace_path"`
}

type LoggingConfig struct {
	Level         string `json:"level"`
	S3Enabled     bool   `json:"s3_enabled"`
	StdoutEnabled bool   `json:"stdout_enabled"`
}

type Config struct {
	MCPServers []MCPServerConfig `json:"mcp_servers"`
	Execution  ExecutionConfig   `json:"execution"`
	Logging    LoggingConfig     `json:"logging"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

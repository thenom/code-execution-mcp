package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"code-execution-mcp/config"
)

// HTTPClientManager manages HTTP connections to backend MCP servers
type HTTPClientManager struct {
	clients map[string]*http.Client
	configs map[string]*config.MCPServerConfig
}

// NewHTTPClientManager creates a new HTTP client manager
func NewHTTPClientManager(serverConfigs []config.MCPServerConfig) *HTTPClientManager {
	manager := &HTTPClientManager{
		clients: make(map[string]*http.Client),
		configs: make(map[string]*config.MCPServerConfig),
	}

	for i := range serverConfigs {
		cfg := &serverConfigs[i]
		manager.configs[cfg.Name] = cfg
		manager.clients[cfg.Name] = &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		}
	}

	return manager
}

// DiscoverTools retrieves available tools from a backend MCP server
func (m *HTTPClientManager) DiscoverTools(ctx context.Context, serverName string) ([]config.Tool, error) {
	cfg, ok := m.configs[serverName]
	if !ok {
		return nil, fmt.Errorf("server %s not found in configuration", serverName)
	}

	client := m.clients[serverName]
	url := fmt.Sprintf("%s/mcp/tools", cfg.URL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to discover tools: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var tools []config.Tool
	if err := json.NewDecoder(resp.Body).Decode(&tools); err != nil {
		return nil, fmt.Errorf("failed to decode tools: %w", err)
	}

	return tools, nil
}

// CallTool executes a tool on a backend MCP server
func (m *HTTPClientManager) CallTool(ctx context.Context, serverName, toolName string, params map[string]interface{}) (interface{}, error) {
	cfg, ok := m.configs[serverName]
	if !ok {
		return nil, fmt.Errorf("server %s not found in configuration", serverName)
	}

	client := m.clients[serverName]
	url := fmt.Sprintf("%s/mcp/call", cfg.URL)

	requestBody := map[string]interface{}{
		"tool":   toolName,
		"params": params,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call tool: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode result: %w", err)
	}

	return result, nil
}

// HealthCheck verifies if a backend MCP server is available
func (m *HTTPClientManager) HealthCheck(ctx context.Context, serverName string) error {
	cfg, ok := m.configs[serverName]
	if !ok {
		return fmt.Errorf("server %s not found in configuration", serverName)
	}

	client := m.clients[serverName]
	url := fmt.Sprintf("%s/health", cfg.URL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned unhealthy status: %d", resp.StatusCode)
	}

	return nil
}

// GetAllServers returns the names of all configured backend servers
func (m *HTTPClientManager) GetAllServers() []string {
	servers := make([]string, 0, len(m.configs))
	for name := range m.configs {
		servers = append(servers, name)
	}
	return servers
}

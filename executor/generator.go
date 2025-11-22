package executor

import (
	"code-execution-mcp/config"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// ModuleGenerator handles the generation of Python modules for MCP tools
type ModuleGenerator struct {
	clientTemplate string
}

// NewModuleGenerator creates a new module generator
func NewModuleGenerator() (*ModuleGenerator, error) {
	// Read the client template
	// Assuming the binary is run from the root directory
	content, err := os.ReadFile("executor/templates/client.py")
	if err != nil {
		return nil, fmt.Errorf("failed to read client template: %w", err)
	}

	return &ModuleGenerator{
		clientTemplate: string(content),
	}, nil
}

// Generate creates the Python module structure in the given workspace
func (mg *ModuleGenerator) Generate(workspacePath string, servers map[string][]config.Tool, serverConfigs []config.MCPServerConfig) error {
	// 1. Create servers directory
	serversDir := filepath.Join(workspacePath, "servers")
	if err := os.MkdirAll(serversDir, 0755); err != nil {
		return fmt.Errorf("failed to create servers directory: %w", err)
	}

	// Create servers/__init__.py
	if err := os.WriteFile(filepath.Join(serversDir, "__init__.py"), []byte(""), 0644); err != nil {
		return fmt.Errorf("failed to create servers/__init__.py: %w", err)
	}

	// 2. Generate _mcp_client.py
	if err := mg.generateClient(workspacePath, serverConfigs); err != nil {
		return fmt.Errorf("failed to generate client: %w", err)
	}

	// 3. Generate modules for each server
	for serverName, tools := range servers {
		if err := mg.generateServerModule(serversDir, serverName, tools); err != nil {
			return fmt.Errorf("failed to generate module for server %s: %w", serverName, err)
		}
	}

	return nil
}

func (mg *ModuleGenerator) generateClient(workspacePath string, serverConfigs []config.MCPServerConfig) error {
	// Create a map of server URLs
	urls := make(map[string]string)
	for _, cfg := range serverConfigs {
		urls[cfg.Name] = cfg.URL
	}

	urlsJSON, err := json.MarshalIndent(urls, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal server URLs: %w", err)
	}

	// Inject URLs into the template
	content := strings.Replace(mg.clientTemplate, "SERVER_URLS = {}", fmt.Sprintf("SERVER_URLS = %s", string(urlsJSON)), 1)

	serversDir := filepath.Join(workspacePath, "servers")
	clientPath := filepath.Join(serversDir, "_mcp_client.py")
	if err := os.WriteFile(clientPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write _mcp_client.py: %w", err)
	}

	return nil
}

func (mg *ModuleGenerator) generateServerModule(serversDir, serverName string, tools []config.Tool) error {
	serverDir := filepath.Join(serversDir, serverName)
	if err := os.MkdirAll(serverDir, 0755); err != nil {
		return fmt.Errorf("failed to create server directory: %w", err)
	}

	// Create __init__.py
	initPath := filepath.Join(serverDir, "__init__.py")
	if err := os.WriteFile(initPath, []byte(""), 0644); err != nil {
		return fmt.Errorf("failed to create __init__.py: %w", err)
	}

	// Generate tool files
	for _, tool := range tools {
		if err := mg.generateToolFile(serverDir, serverName, tool); err != nil {
			return fmt.Errorf("failed to generate tool file for %s: %w", tool.Name, err)
		}

		// Append import to __init__.py
		importStmt := fmt.Sprintf("from .%s import %s\n", tool.Name, tool.Name)
		f, err := os.OpenFile(initPath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to append to __init__.py: %w", err)
		}
		if _, err := f.WriteString(importStmt); err != nil {
			f.Close()
			return err
		}
		f.Close()
	}

	return nil
}

const toolTemplate = `from .._mcp_client import call_mcp_tool

def {{.ToolName}}(**kwargs):
    """
    {{.Description}}
    """
    return call_mcp_tool("{{.ServerName}}", "{{.ToolName}}", kwargs)
`

func (mg *ModuleGenerator) generateToolFile(serverDir, serverName string, tool config.Tool) error {
	tmpl, err := template.New("tool").Parse(toolTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse tool template: %w", err)
	}

	filePath := filepath.Join(serverDir, fmt.Sprintf("%s.py", tool.Name))
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create tool file: %w", err)
	}
	defer f.Close()

	data := struct {
		ToolName    string
		ServerName  string
		Description string
	}{
		ToolName:    tool.Name,
		ServerName:  serverName,
		Description: tool.Description,
	}

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("failed to execute tool template: %w", err)
	}

	return nil
}

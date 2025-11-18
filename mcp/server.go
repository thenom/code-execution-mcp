package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"code-execution-mcp/config"
	"code-execution-mcp/executor"
)

type Server struct {
	mcpServer  *server.MCPServer
	config     *config.Config
	httpClient *HTTPClientManager
	executor   *executor.PythonExecutor
}

func NewServer(cfg *config.Config) (*Server, error) {
	s := server.NewMCPServer(
		"code-execution-server",
		"0.1.0",
	)

	// Initialize HTTP client manager
	httpClient := NewHTTPClientManager(cfg.MCPServers)

	// Initialize Python executor
	exec, err := executor.NewPythonExecutor(&cfg.Execution)
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}

	return &Server{
		mcpServer:  s,
		config:     cfg,
		httpClient: httpClient,
		executor:   exec,
	}, nil
}

func (s *Server) Start() error {
	// Register tools
	s.mcpServer.AddTool(mcp.NewTool("execute_code",
		mcp.WithDescription("Execute Python code with access to MCP tools"),
		mcp.WithString("code", mcp.Required(), mcp.Description("Python code to execute")),
		mcp.WithNumber("timeout", mcp.Description("Execution timeout in seconds")),
	), s.executeCodeHandler)

	s.mcpServer.AddTool(mcp.NewTool("search_tools",
		mcp.WithDescription("Search available MCP tools with filtering options"),
		mcp.WithString("query", mcp.Description("Search term for tool names and descriptions")),
		mcp.WithString("server_filter", mcp.Description("Filter by server name (exact or partial match)")),
		mcp.WithString("detail_level", mcp.Description("Level of detail to return (name/description/full)")),
	), s.searchToolsHandler)

	// Start the server using STDIO
	return server.ServeStdio(s.mcpServer)
}

func (s *Server) executeCodeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, ok := request.Arguments["code"].(string)
	if !ok {
		return mcp.NewToolResultError("code argument is required and must be a string"), nil
	}

	// Get timeout if provided
	timeout := 0
	if t, ok := request.Arguments["timeout"].(float64); ok {
		timeout = int(t)
	}

	// Execute the code
	result, err := s.executor.Execute(ctx, code, timeout)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Execution failed: %v", err)), nil
	}

	// Format the result
	var output string
	if result.Error != "" {
		output = fmt.Sprintf("Exit Code: %d\n\nError:\n%s\n\nOutput:\n%s",
			result.ExitCode, result.Error, result.Output)
	} else {
		output = result.Output
	}

	return mcp.NewToolResultText(output), nil
}

func (s *Server) searchToolsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, _ := request.Arguments["query"].(string)
	serverFilter, _ := request.Arguments["server_filter"].(string)
	detailLevel, _ := request.Arguments["detail_level"].(string)

	if detailLevel == "" {
		detailLevel = "description"
	}

	// Collect tools from all backend servers
	var allTools []map[string]interface{}

	servers := s.httpClient.GetAllServers()
	for _, serverName := range servers {
		// Apply server filter if provided
		if serverFilter != "" && !strings.Contains(strings.ToLower(serverName), strings.ToLower(serverFilter)) {
			continue
		}

		tools, err := s.httpClient.DiscoverTools(ctx, serverName)
		if err != nil {
			// Log error but continue with other servers
			continue
		}

		for _, tool := range tools {
			// Apply query filter if provided
			if query != "" {
				queryLower := strings.ToLower(query)
				nameLower := strings.ToLower(tool.Name)
				descLower := strings.ToLower(tool.Description)

				if !strings.Contains(nameLower, queryLower) && !strings.Contains(descLower, queryLower) {
					continue
				}
			}

			// Format based on detail level
			toolInfo := make(map[string]interface{})
			toolInfo["server"] = serverName
			toolInfo["name"] = tool.Name

			if detailLevel == "description" || detailLevel == "full" {
				toolInfo["description"] = tool.Description
			}

			if detailLevel == "full" {
				toolInfo["inputSchema"] = tool.InputSchema
			}

			allTools = append(allTools, toolInfo)
		}
	}

	// Format response
	resultJSON, err := json.MarshalIndent(allTools, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format results: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultJSON)), nil
}

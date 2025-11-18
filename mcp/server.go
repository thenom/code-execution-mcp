package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"code-execution-mcp/config"
)

type Server struct {
	mcpServer *server.MCPServer
	config    *config.Config
}

func NewServer(cfg *config.Config) *Server {
	s := server.NewMCPServer(
		cfg.Execution.WorkspacePath, // Using workspace path as name for now, or just "code-execution-server"
		"0.1.0",
	)

	return &Server{
		mcpServer: s,
		config:    cfg,
	}
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
		mcp.WithString("detail_level", mcp.Description("Level of detail to return")),
	), s.searchToolsHandler)

	// Start the server using STDIO
	return server.ServeStdio(s.mcpServer)
}

func (s *Server) executeCodeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, ok := request.Arguments["code"].(string)
	if !ok {
		return mcp.NewToolResultError("code argument is required and must be a string"), nil
	}

	// TODO: Implement actual execution logic
	return mcp.NewToolResultText(fmt.Sprintf("Executed code: %s", code)), nil
}

func (s *Server) searchToolsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// TODO: Implement actual search logic
	return mcp.NewToolResultText("Search results placeholder"), nil
}

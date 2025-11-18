package main

import (
	"flag"
	"log"

	"code-execution-mcp/config"
	"code-execution-mcp/mcp"
)

func main() {
	configPath := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize MCP server
	server, err := mcp.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Start the server (blocks until shutdown)
	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

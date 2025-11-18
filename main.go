package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"code-execution-mcp/config"
)

func main() {
	configPath := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// TODO: Initialize MCP server and other components
	fmt.Printf("Starting MCP Code Execution Server with config: %+v\n", cfg)
	
	// Keep the process running for now
	select {}
}

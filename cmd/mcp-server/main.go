package main

import (
	"log"

	"github.com/divesh/contextforge/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Create the MCP server
	mcpServer := mcp.NewContextForgeServer()

	log.Println("Starting ContextForge MCP Server...")

	// Start the server with stdio transport
	if err := server.ServeStdio(mcpServer); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

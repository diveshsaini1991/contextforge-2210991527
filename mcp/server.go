package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewContextForgeServer creates a new MCP server for ContextForge
func NewContextForgeServer() *server.MCPServer {
	s := server.NewMCPServer(
		"ContextForge",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Register tools
	registerTools(s)

	return s
}

// registerTools registers all MCP tools
func registerTools(s *server.MCPServer) {
	// Tool 1: build_repo_context
	s.AddTool(mcp.NewTool("build_repo_context",
		mcp.WithDescription("Scans repository and creates comprehensive context with all functions and existing tests. Saves to .contextforge/context.json"),
		mcp.WithString("repo_path",
			mcp.Required(),
			mcp.Description("Path to the Go repository to analyze"),
		),
	), HandleBuildRepoContext)

	// Tool 2: analyze_test_scenarios
	s.AddTool(mcp.NewTool("analyze_test_scenarios",
		mcp.WithDescription("Generates list of all test scenarios that should exist (happy path, error cases, edge cases, boundary). Saves to .contextforge/scenarios.json. Uses existing context if available."),
		mcp.WithString("repo_path",
			mcp.Required(),
			mcp.Description("Path to the Go repository"),
		),
	), HandleAnalyzeTestScenarios)

	// Tool 3: check_test_coverage
	s.AddTool(mcp.NewTool("check_test_coverage",
		mcp.WithDescription("Compares test scenarios against existing tests and generates coverage report. Saves to .contextforge/coverage_report.json. Uses existing context and scenarios if available."),
		mcp.WithString("repo_path",
			mcp.Required(),
			mcp.Description("Path to the Go repository"),
		),
	), HandleCheckTestCoverage)

	// Tool 4: generate_test_stubs
	s.AddTool(mcp.NewTool("generate_test_stubs",
		mcp.WithDescription("Creates test files with stub functions for missing tests. Each stub prints 'test not yet built' and can be implemented later."),
		mcp.WithString("repo_path",
			mcp.Required(),
			mcp.Description("Path to the Go repository"),
		),
	), HandleGenerateTestStubs)

	// Tool 5: get_context
	s.AddTool(mcp.NewTool("get_context",
		mcp.WithDescription("Retrieves previously saved repository context from .contextforge/context.json"),
		mcp.WithString("repo_path",
			mcp.Required(),
			mcp.Description("Path to the Go repository"),
		),
	), HandleGetContext)
}

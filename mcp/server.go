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
	// Tool 1: analyze_repository
	s.AddTool(mcp.NewTool("analyze_repository",
		mcp.WithDescription("Analyzes a Go repository and extracts function-level metadata"),
		mcp.WithString("repo_path",
			mcp.Required(),
			mcp.Description("Path to the Go repository to analyze"),
		),
	), HandleAnalyzeRepository)

	// Tool 2: analyze_coverage
	s.AddTool(mcp.NewTool("analyze_coverage",
		mcp.WithDescription("Runs coverage analysis and identifies gaps in test coverage"),
		mcp.WithString("repo_path",
			mcp.Required(),
			mcp.Description("Path to the Go repository to analyze"),
		),
	), HandleAnalyzeCoverage)

	// Tool 3: get_test_recommendations
	s.AddTool(mcp.NewTool("get_test_recommendations",
		mcp.WithDescription("Returns prioritized list of uncovered functions that need tests"),
		mcp.WithString("repo_path",
			mcp.Required(),
			mcp.Description("Path to the Go repository"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of recommendations (default: 10)"),
		),
	), HandleGetTestRecommendations)

	// Tool 4: generate_unit_tests
	s.AddTool(mcp.NewTool("generate_unit_tests",
		mcp.WithDescription("Extracts context for AI-assisted unit test generation. Returns function source code, imports, type definitions, and detailed instructions. The AI assistant should use this context to generate and write comprehensive unit tests."),
		mcp.WithString("repo_path",
			mcp.Required(),
			mcp.Description("Path to the Go repository"),
		),
		mcp.WithArray("function_ids",
			mcp.Required(),
			mcp.Description("List of function identifiers to generate tests for (format: 'FunctionName' or 'package.FunctionName')"),
		),
	), HandleGenerateUnitTests)

	// Tool 5: verify_coverage_improvement
	s.AddTool(mcp.NewTool("verify_coverage_improvement",
		mcp.WithDescription("Re-runs coverage analysis to verify improvements after test generation"),
		mcp.WithString("repo_path",
			mcp.Required(),
			mcp.Description("Path to the Go repository"),
		),
	), HandleVerifyCoverageImprovement)
}

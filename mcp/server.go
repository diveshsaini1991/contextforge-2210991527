package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewContextForgeServer creates a new MCP server for ContextForge.
func NewContextForgeServer() *server.MCPServer {
	s := server.NewMCPServer(
		"ContextForge",
		"1.1.0",
		server.WithToolCapabilities(true),
		server.WithRecovery(),
		server.WithInstructions(
			"ContextForge analyzes Go repositories for test coverage gaps. "+
				"Provide either repo_path (local) or repo_url (GitHub/GitLab URL, cloned automatically). "+
				"Recommended workflow: "+
				"1) build_repo_context -> 2) analyze_test_scenarios -> 3) check_test_coverage -> 4) generate_test_stubs. "+
				"Each step caches results in .contextforge/ for reuse. "+
				"Use get_context, get_scenarios, or get_coverage_report to retrieve cached results.",
		),
	)

	registerTools(s)
	return s
}

func registerTools(s *server.MCPServer) {
	s.AddTool(mcp.NewTool("build_repo_context",
		mcp.WithDescription(
			"Step 1 of 4. Scans a Go repository and builds comprehensive context: "+
				"every function with signature, complexity, line count, and existing test mappings. "+
				"Saves to .contextforge/context.json. Run this first before other tools. "+
				"Provide repo_path (local) or repo_url (remote, auto-cloned). "+
				"Returns JSON with summary stats, package list, and all function details.",
		),
		repoPathParam(),
		repoURLParam(),
		forceParam(),
		packageFilterParam(),
	), HandleBuildRepoContext)

	s.AddTool(mcp.NewTool("analyze_test_scenarios",
		mcp.WithDescription(
			"Step 2 of 4. Generates test scenarios that should exist for every function: "+
				"happy path, error cases, edge cases, and boundary conditions. "+
				"Saves to .contextforge/scenarios.json. "+
				"Auto-builds context first if not cached (or use force=true to rebuild). "+
				"Returns JSON with all scenarios and suggested test function names.",
		),
		repoPathParam(),
		repoURLParam(),
		forceParam(),
		packageFilterParam(),
	), HandleAnalyzeTestScenarios)

	s.AddTool(mcp.NewTool("check_test_coverage",
		mcp.WithDescription(
			"Step 3 of 4. Compares generated scenarios against existing tests and "+
				"produces a gap analysis report with per-package coverage percentages. "+
				"Saves to .contextforge/coverage_report.json. "+
				"Auto-builds context and scenarios if not cached. "+
				"Returns JSON with coverage stats, package breakdown, and list of missing tests.",
		),
		repoPathParam(),
		repoURLParam(),
		forceParam(),
		packageFilterParam(),
	), HandleCheckTestCoverage)

	s.AddTool(mcp.NewTool("generate_test_stubs",
		mcp.WithDescription(
			"Step 4 of 4. Creates test files with stub functions for every missing test "+
				"identified by check_test_coverage. Each stub calls t.Skip() with a description. "+
				"Appends to existing test files or creates new ones. Non-destructive. "+
				"Requires: context and coverage report must exist (run steps 1-3 first). "+
				"Returns summary with list of created/modified test files.",
		),
		repoPathParam(),
		repoURLParam(),
	), HandleGenerateTestStubs)

	s.AddTool(mcp.NewTool("get_context",
		mcp.WithDescription(
			"Retrieves saved repository context from .contextforge/context.json without re-scanning. "+
				"Returns error if build_repo_context has not been run yet.",
		),
		repoPathParam(),
		repoURLParam(),
	), HandleGetContext)

	s.AddTool(mcp.NewTool("get_scenarios",
		mcp.WithDescription(
			"Retrieves saved test scenario analysis from .contextforge/scenarios.json without re-analyzing. "+
				"Returns error if analyze_test_scenarios has not been run yet.",
		),
		repoPathParam(),
		repoURLParam(),
	), HandleGetScenarios)

	s.AddTool(mcp.NewTool("get_coverage_report",
		mcp.WithDescription(
			"Retrieves saved coverage report from .contextforge/coverage_report.json without re-analyzing. "+
				"Returns error if check_test_coverage has not been run yet.",
		),
		repoPathParam(),
		repoURLParam(),
	), HandleGetCoverageReport)
}

func repoPathParam() mcp.ToolOption {
	return mcp.WithString("repo_path",
		mcp.Description("Absolute path to a local Go repository. Provide this OR repo_url."),
	)
}

func repoURLParam() mcp.ToolOption {
	return mcp.WithString("repo_url",
		mcp.Description("Git clone URL (e.g. https://github.com/user/repo). Auto-cloned and cached. Provide this OR repo_path."),
	)
}

func forceParam() mcp.ToolOption {
	return mcp.WithBoolean("force",
		mcp.Description("Skip cached results and force a fresh scan/analysis. Also re-clones if using repo_url. Default: false"),
	)
}

func packageFilterParam() mcp.ToolOption {
	return mcp.WithString("package_filter",
		mcp.Description("Only process packages whose path contains this substring. Example: 'pkg/api'"),
	)
}

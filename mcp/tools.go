package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/divesh/contextforge/internal/builder"
	"github.com/mark3labs/mcp-go/mcp"
)

// Tool handlers for MCP server

// BuildRepoContextInput defines input for build_repo_context tool
type BuildRepoContextInput struct {
	RepoPath string `json:"repo_path" jsonschema:"required,description=Path to the Go repository"`
}

// AnalyzeTestScenariosInput defines input for analyze_test_scenarios tool
type AnalyzeTestScenariosInput struct {
	RepoPath string `json:"repo_path" jsonschema:"required,description=Path to the Go repository"`
}

// CheckTestCoverageInput defines input for check_test_coverage tool
type CheckTestCoverageInput struct {
	RepoPath string `json:"repo_path" jsonschema:"required,description=Path to the Go repository"`
}

// GenerateTestStubsInput defines input for generate_test_stubs tool
type GenerateTestStubsInput struct {
	RepoPath string `json:"repo_path" jsonschema:"required,description=Path to the Go repository"`
}

// GetContextInput defines input for get_context tool
type GetContextInput struct {
	RepoPath string `json:"repo_path" jsonschema:"required,description=Path to the Go repository"`
}

// HandleBuildRepoContext handles the build_repo_context tool
func HandleBuildRepoContext(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments type"), nil
	}

	var input BuildRepoContextInput
	if err := mapToStruct(args, &input); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid input: %v", err)), nil
	}

	// Build comprehensive repository context
	contextBuilder := builder.NewContextBuilder(input.RepoPath)
	repoContext, err := contextBuilder.BuildContext()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to build context: %v", err)), nil
	}

	// Convert to JSON
	result, err := json.MarshalIndent(repoContext, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(result)), nil
}

// HandleAnalyzeTestScenarios handles the analyze_test_scenarios tool
func HandleAnalyzeTestScenarios(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments type"), nil
	}

	var input AnalyzeTestScenariosInput
	if err := mapToStruct(args, &input); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid input: %v", err)), nil
	}

	// Try to load existing context
	contextBuilder := builder.NewContextBuilder(input.RepoPath)
	repoContext, err := contextBuilder.LoadContext()
	if err != nil {
		// Context doesn't exist, build it
		repoContext, err = contextBuilder.BuildContext()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to build context: %v", err)), nil
		}
	}

	// Analyze test scenarios
	scenarioAnalyzer := builder.NewScenarioAnalyzer(input.RepoPath)
	scenarios, err := scenarioAnalyzer.AnalyzeScenarios(repoContext)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to analyze scenarios: %v", err)), nil
	}

	// Convert to JSON
	result, err := json.MarshalIndent(scenarios, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(result)), nil
}

// HandleCheckTestCoverage handles the check_test_coverage tool
func HandleCheckTestCoverage(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments type"), nil
	}

	var input CheckTestCoverageInput
	if err := mapToStruct(args, &input); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid input: %v", err)), nil
	}

	// Load or build context
	contextBuilder := builder.NewContextBuilder(input.RepoPath)
	repoContext, err := contextBuilder.LoadContext()
	if err != nil {
		repoContext, err = contextBuilder.BuildContext()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to build context: %v", err)), nil
		}
	}

	// Load or analyze scenarios
	scenarioAnalyzer := builder.NewScenarioAnalyzer(input.RepoPath)
	scenarios, err := scenarioAnalyzer.LoadAnalysis()
	if err != nil {
		scenarios, err = scenarioAnalyzer.AnalyzeScenarios(repoContext)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to analyze scenarios: %v", err)), nil
		}
	}

	// Check coverage
	coverageChecker := builder.NewCoverageChecker(input.RepoPath)
	report, err := coverageChecker.CheckCoverage(repoContext, scenarios)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to check coverage: %v", err)), nil
	}

	// Convert to JSON
	result, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(result)), nil
}

// HandleGenerateTestStubs handles the generate_test_stubs tool
func HandleGenerateTestStubs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments type"), nil
	}

	var input GenerateTestStubsInput
	if err := mapToStruct(args, &input); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid input: %v", err)), nil
	}

	// Load context
	contextBuilder := builder.NewContextBuilder(input.RepoPath)
	repoContext, err := contextBuilder.LoadContext()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Context not found. Run build_repo_context first: %v", err)), nil
	}

	// Load coverage report
	coverageChecker := builder.NewCoverageChecker(input.RepoPath)
	report, err := coverageChecker.LoadReport()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Coverage report not found. Run check_test_coverage first: %v", err)), nil
	}

	// Generate stubs
	stubGenerator := builder.NewStubGenerator(input.RepoPath)
	createdFiles, err := stubGenerator.GenerateStubs(repoContext, report)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to generate stubs: %v", err)), nil
	}

	// Return summary
	summary := map[string]interface{}{
		"created_files":     createdFiles,
		"total_stubs":       len(report.MissingTests),
		"message":           fmt.Sprintf("Generated %d test stubs in %d file(s)", len(report.MissingTests), len(createdFiles)),
		"next_steps":        "Implement the test logic in the generated stub functions",
	}

	result, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(result)), nil
}

// HandleGetContext handles the get_context tool
func HandleGetContext(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments type"), nil
	}

	var input GetContextInput
	if err := mapToStruct(args, &input); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid input: %v", err)), nil
	}

	// Load context
	contextBuilder := builder.NewContextBuilder(input.RepoPath)
	repoContext, err := contextBuilder.LoadContext()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Context not found. Run build_repo_context first: %v", err)), nil
	}

	// Convert to JSON
	result, err := json.MarshalIndent(repoContext, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(result)), nil
}

// mapToStruct converts a map to a struct
func mapToStruct(m map[string]interface{}, result interface{}) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, result)
}

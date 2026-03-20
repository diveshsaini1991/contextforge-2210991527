package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/divesh/contextforge/internal/coverage"
	"github.com/divesh/contextforge/internal/generator"
	"github.com/divesh/contextforge/internal/models"
	"github.com/divesh/contextforge/internal/scanner"
	"github.com/mark3labs/mcp-go/mcp"
)

// Tool handlers for MCP server

// AnalyzeRepositoryInput defines input for analyze_repository tool
type AnalyzeRepositoryInput struct {
	RepoPath string `json:"repo_path" jsonschema:"required,description=Path to the Go repository"`
}

// AnalyzeCoverageInput defines input for analyze_coverage tool
type AnalyzeCoverageInput struct {
	RepoPath string `json:"repo_path" jsonschema:"required,description=Path to the Go repository"`
}

// GetTestRecommendationsInput defines input for get_test_recommendations tool
type GetTestRecommendationsInput struct {
	RepoPath string `json:"repo_path" jsonschema:"required,description=Path to the Go repository"`
	Limit    int    `json:"limit" jsonschema:"description=Maximum number of recommendations to return"`
}

// GenerateUnitTestsInput defines input for generate_unit_tests tool
type GenerateUnitTestsInput struct {
	RepoPath    string   `json:"repo_path" jsonschema:"required,description=Path to the Go repository"`
	FunctionIDs []string `json:"function_ids" jsonschema:"required,description=List of function identifiers to generate tests for"`
}

// VerifyCoverageImprovementInput defines input for verify_coverage_improvement tool
type VerifyCoverageImprovementInput struct {
	RepoPath string `json:"repo_path" jsonschema:"required,description=Path to the Go repository"`
}

// HandleAnalyzeRepository handles the analyze_repository tool
func HandleAnalyzeRepository(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments type"), nil
	}

	var input AnalyzeRepositoryInput
	if err := mapToStruct(args, &input); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid input: %v", err)), nil
	}

	// Scan the repository
	repoCtx, err := scanner.ScanRepository(input.RepoPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to scan repository: %v", err)), nil
	}

	// Convert to JSON
	result, err := json.MarshalIndent(repoCtx, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(result)), nil
}

// HandleAnalyzeCoverage handles the analyze_coverage tool
func HandleAnalyzeCoverage(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments type"), nil
	}

	var input AnalyzeCoverageInput
	if err := mapToStruct(args, &input); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid input: %v", err)), nil
	}

	// First, analyze the repository
	repoCtx, err := scanner.ScanRepository(input.RepoPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to scan repository: %v", err)), nil
	}

	// Then analyze coverage
	analyzer := coverage.NewAnalyzer(input.RepoPath, repoCtx)
	report, err := analyzer.Analyze()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to analyze coverage: %v", err)), nil
	}

	// Convert to JSON
	result, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(result)), nil
}

// HandleGetTestRecommendations handles the get_test_recommendations tool
func HandleGetTestRecommendations(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments type"), nil
	}

	var input GetTestRecommendationsInput
	if err := mapToStruct(args, &input); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid input: %v", err)), nil
	}

	if input.Limit == 0 {
		input.Limit = 10 // Default limit
	}

	// Analyze repository and coverage
	repoCtx, err := scanner.ScanRepository(input.RepoPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to scan repository: %v", err)), nil
	}

	analyzer := coverage.NewAnalyzer(input.RepoPath, repoCtx)
	report, err := analyzer.Analyze()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to analyze coverage: %v", err)), nil
	}

	// Collect all uncovered functions
	var uncovered []models.FunctionCoverage
	for _, pkg := range report.CoverageByPackage {
		uncovered = append(uncovered, pkg.UncoveredFunctions...)
	}

	// Prioritize and get top N
	prioritizer := generator.NewPrioritizer()
	recommendations := prioritizer.GetTopPriority(uncovered, input.Limit)

	// Convert to JSON
	result, err := json.MarshalIndent(recommendations, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(result)), nil
}

// HandleGenerateUnitTests handles the generate_unit_tests tool
// Returns context for the AI assistant to generate tests
func HandleGenerateUnitTests(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments type"), nil
	}

	var input GenerateUnitTestsInput
	if err := mapToStruct(args, &input); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid input: %v", err)), nil
	}

	// Analyze repository and coverage
	repoCtx, err := scanner.ScanRepository(input.RepoPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to scan repository: %v", err)), nil
	}

	analyzer := coverage.NewAnalyzer(input.RepoPath, repoCtx)
	report, err := analyzer.Analyze()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to analyze coverage: %v", err)), nil
	}

	// Find functions matching the IDs
	var targetFunctions []models.FunctionCoverage
	for _, pkg := range report.CoverageByPackage {
		for _, fn := range pkg.UncoveredFunctions {
			for _, id := range input.FunctionIDs {
				if fn.Name == id || fmt.Sprintf("%s.%s", pkg.Package, fn.Name) == id {
					targetFunctions = append(targetFunctions, fn)
				}
			}
		}
	}

	if len(targetFunctions) == 0 {
		return mcp.NewToolResultError("No matching functions found"), nil
	}

	// Prepare test generation context (instead of generating tests directly)
	gen := generator.NewTestGenerator(input.RepoPath)
	genResult, err := gen.PrepareTestGeneration(targetFunctions)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to prepare test generation: %v", err)), nil
	}

	// Return comprehensive context for AI to generate tests
	result, err := json.MarshalIndent(genResult, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(result)), nil
}

// HandleVerifyCoverageImprovement handles the verify_coverage_improvement tool
func HandleVerifyCoverageImprovement(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments type"), nil
	}

	var input VerifyCoverageImprovementInput
	if err := mapToStruct(args, &input); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid input: %v", err)), nil
	}

	// Analyze current coverage
	repoCtx, err := scanner.ScanRepository(input.RepoPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to scan repository: %v", err)), nil
	}

	analyzer := coverage.NewAnalyzer(input.RepoPath, repoCtx)
	report, err := analyzer.Analyze()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to analyze coverage: %v", err)), nil
	}

	// Return coverage summary
	summary := map[string]interface{}{
		"overall_coverage":        report.OverallCoverage,
		"total_functions":         report.TotalUncovered + report.TotalPartiallyCovered + report.TotalFullyCovered,
		"uncovered_functions":     report.TotalUncovered,
		"partially_covered":       report.TotalPartiallyCovered,
		"fully_covered_functions": report.TotalFullyCovered,
		"analyzed_at":             report.AnalyzedAt,
	}

	result, err := json.MarshalIndent(summary, "", "  ")
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

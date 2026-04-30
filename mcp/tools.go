package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/divesh/contextforge/internal/builder"
	"github.com/divesh/contextforge/internal/models"
	"github.com/mark3labs/mcp-go/mcp"
)

// ToolInput is the shared input for all ContextForge tools.
type ToolInput struct {
	RepoPath      string `json:"repo_path"`
	Force         bool   `json:"force"`
	PackageFilter string `json:"package_filter"`
}

// parseAndValidate extracts ToolInput from the request and validates repo_path.
func parseAndValidate(request mcp.CallToolRequest) (*ToolInput, *mcp.CallToolResult) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return nil, mcp.NewToolResultError("invalid arguments: expected JSON object")
	}

	var input ToolInput
	if err := mapToStruct(args, &input); err != nil {
		return nil, mcp.NewToolResultErrorf("invalid input: %v", err)
	}

	if input.RepoPath == "" {
		return nil, mcp.NewToolResultError("repo_path is required")
	}

	info, err := os.Stat(input.RepoPath)
	if err != nil {
		return nil, mcp.NewToolResultErrorf("repo_path %q does not exist: %v", input.RepoPath, err)
	}
	if !info.IsDir() {
		return nil, mcp.NewToolResultErrorf("repo_path %q is not a directory", input.RepoPath)
	}

	if _, err := os.Stat(filepath.Join(input.RepoPath, "go.mod")); err != nil {
		return nil, mcp.NewToolResultErrorf("repo_path %q does not contain a go.mod file", input.RepoPath)
	}

	return &input, nil
}

func resultJSON(v any) (*mcp.CallToolResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return mcp.NewToolResultErrorf("failed to marshal result: %v", err), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}

func newContextBuilder(input *ToolInput) *builder.ContextBuilder {
	cb := builder.NewContextBuilder(input.RepoPath)
	if input.PackageFilter != "" {
		cb.SetPackageFilter(input.PackageFilter)
	}
	return cb
}

func newScenarioAnalyzer(input *ToolInput) *builder.ScenarioAnalyzer {
	sa := builder.NewScenarioAnalyzer(input.RepoPath)
	if input.PackageFilter != "" {
		sa.SetPackageFilter(input.PackageFilter)
	}
	return sa
}

// HandleBuildRepoContext scans the repository and builds comprehensive context.
func HandleBuildRepoContext(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input, errResult := parseAndValidate(request)
	if errResult != nil {
		return errResult, nil
	}

	repoContext, err := newContextBuilder(input).BuildContext(ctx)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to build context: %v", err), nil
	}

	return resultJSON(repoContext)
}

// HandleAnalyzeTestScenarios generates test scenarios for all functions.
func HandleAnalyzeTestScenarios(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input, errResult := parseAndValidate(request)
	if errResult != nil {
		return errResult, nil
	}

	repoContext, err := loadOrBuildContext(ctx, input)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to get context: %v", err), nil
	}

	scenarios, err := newScenarioAnalyzer(input).AnalyzeScenarios(ctx, repoContext)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to analyze scenarios: %v", err), nil
	}

	return resultJSON(scenarios)
}

// HandleCheckTestCoverage compares scenarios against existing tests.
func HandleCheckTestCoverage(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input, errResult := parseAndValidate(request)
	if errResult != nil {
		return errResult, nil
	}

	repoContext, err := loadOrBuildContext(ctx, input)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to get context: %v", err), nil
	}

	scenarioAnalyzer := newScenarioAnalyzer(input)
	var scenarios *models.ScenarioAnalysis

	if !input.Force {
		scenarios, _ = scenarioAnalyzer.LoadAnalysis(ctx)
	}
	if scenarios == nil {
		scenarios, err = scenarioAnalyzer.AnalyzeScenarios(ctx, repoContext)
		if err != nil {
			return mcp.NewToolResultErrorf("failed to analyze scenarios: %v", err), nil
		}
	}

	report, err := builder.NewCoverageChecker(input.RepoPath).CheckCoverage(ctx, repoContext, scenarios)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to check coverage: %v", err), nil
	}

	return resultJSON(report)
}

// HandleGenerateTestStubs creates test stub files for missing tests.
func HandleGenerateTestStubs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input, errResult := parseAndValidate(request)
	if errResult != nil {
		return errResult, nil
	}

	cb := builder.NewContextBuilder(input.RepoPath)
	repoContext, err := cb.LoadContext(ctx)
	if err != nil {
		return mcp.NewToolResultErrorf("context not found; run build_repo_context first: %v", err), nil
	}

	checker := builder.NewCoverageChecker(input.RepoPath)
	report, err := checker.LoadReport(ctx)
	if err != nil {
		return mcp.NewToolResultErrorf("coverage report not found; run check_test_coverage first: %v", err), nil
	}

	createdFiles, err := builder.NewStubGenerator(input.RepoPath).GenerateStubs(ctx, repoContext, report)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to generate stubs: %v", err), nil
	}

	return resultJSON(map[string]any{
		"created_files": createdFiles,
		"total_stubs":   len(report.MissingTests),
		"message":       fmt.Sprintf("Generated %d test stubs across %d file(s)", len(report.MissingTests), len(createdFiles)),
	})
}

// HandleGetContext retrieves saved repository context.
func HandleGetContext(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input, errResult := parseAndValidate(request)
	if errResult != nil {
		return errResult, nil
	}

	repoContext, err := builder.NewContextBuilder(input.RepoPath).LoadContext(ctx)
	if err != nil {
		return mcp.NewToolResultErrorf("context not found; run build_repo_context first: %v", err), nil
	}

	return resultJSON(repoContext)
}

// HandleGetScenarios retrieves saved test scenario analysis.
func HandleGetScenarios(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input, errResult := parseAndValidate(request)
	if errResult != nil {
		return errResult, nil
	}

	analysis, err := builder.NewScenarioAnalyzer(input.RepoPath).LoadAnalysis(ctx)
	if err != nil {
		return mcp.NewToolResultErrorf("scenarios not found; run analyze_test_scenarios first: %v", err), nil
	}

	return resultJSON(analysis)
}

// HandleGetCoverageReport retrieves saved coverage report.
func HandleGetCoverageReport(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input, errResult := parseAndValidate(request)
	if errResult != nil {
		return errResult, nil
	}

	report, err := builder.NewCoverageChecker(input.RepoPath).LoadReport(ctx)
	if err != nil {
		return mcp.NewToolResultErrorf("coverage report not found; run check_test_coverage first: %v", err), nil
	}

	return resultJSON(report)
}

// loadOrBuildContext loads cached context or builds it fresh.
func loadOrBuildContext(ctx context.Context, input *ToolInput) (*models.RepoContext, error) {
	cb := newContextBuilder(input)

	if !input.Force {
		if repoContext, err := cb.LoadContext(ctx); err == nil {
			return repoContext, nil
		}
	}

	return cb.BuildContext(ctx)
}

func mapToStruct(m map[string]interface{}, result interface{}) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, result)
}

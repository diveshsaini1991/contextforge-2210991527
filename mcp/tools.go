package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/diveshsaini1991/contextforge-2210991527/internal/builder"
	"github.com/diveshsaini1991/contextforge-2210991527/internal/models"
	"github.com/mark3labs/mcp-go/mcp"
)

// ToolInput is the shared input for all ContextForge tools.
type ToolInput struct {
	RepoPath      string `json:"repo_path"`
	RepoURL       string `json:"repo_url"`
	Force         bool   `json:"force"`
	PackageFilter string `json:"package_filter"`
}

var (
	cloneCache   = make(map[string]string)
	cloneCacheMu sync.Mutex
)

// parseAndValidate extracts ToolInput from the request and resolves the repo path.
// Supports both local repo_path and remote repo_url (clones to temp dir).
func parseAndValidate(ctx context.Context, request mcp.CallToolRequest) (*ToolInput, *mcp.CallToolResult) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return nil, mcp.NewToolResultError("invalid arguments: expected JSON object")
	}

	var input ToolInput
	if err := mapToStruct(args, &input); err != nil {
		return nil, mcp.NewToolResultErrorf("invalid input: %v", err)
	}

	if input.RepoPath == "" && input.RepoURL == "" {
		return nil, mcp.NewToolResultError("either repo_path or repo_url is required")
	}

	// If repo_url is provided, clone (or reuse cached clone)
	if input.RepoURL != "" {
		clonePath, err := cloneRepo(ctx, input.RepoURL, input.Force)
		if err != nil {
			return nil, mcp.NewToolResultErrorf("failed to clone %q: %v", input.RepoURL, err)
		}
		input.RepoPath = clonePath
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

// cloneRepo clones a git repo URL to a temp directory.
// Caches clones by URL to avoid re-cloning on subsequent tool calls.
func cloneRepo(ctx context.Context, repoURL string, force bool) (string, error) {
	cloneCacheMu.Lock()
	defer cloneCacheMu.Unlock()

	// Reuse cached clone if available and not forced
	if !force {
		if dir, ok := cloneCache[repoURL]; ok {
			if _, err := os.Stat(dir); err == nil {
				return dir, nil
			}
			delete(cloneCache, repoURL)
		}
	}

	// Sanitize URL into a directory-safe name
	safeName := strings.NewReplacer(
		"https://", "", "http://", "",
		"git@", "", ":", "-",
		"/", "-", ".git", "",
	).Replace(repoURL)

	tmpDir, err := os.MkdirTemp("", "contextforge-"+safeName+"-")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	cmd := exec.CommandContext(ctx, "git", "clone", "--depth=1", "--single-branch", repoURL, tmpDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("git clone failed: %s: %w", strings.TrimSpace(string(output)), err)
	}

	cloneCache[repoURL] = tmpDir
	return tmpDir, nil
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
	input, errResult := parseAndValidate(ctx, request)
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
	input, errResult := parseAndValidate(ctx, request)
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
	input, errResult := parseAndValidate(ctx, request)
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
	input, errResult := parseAndValidate(ctx, request)
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
	input, errResult := parseAndValidate(ctx, request)
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
	input, errResult := parseAndValidate(ctx, request)
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
	input, errResult := parseAndValidate(ctx, request)
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

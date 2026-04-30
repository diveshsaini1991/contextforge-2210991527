package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/divesh/contextforge/internal/models"
)

// ScenarioAnalyzer generates test scenarios for functions.
type ScenarioAnalyzer struct {
	repoPath      string
	contextDir    string
	packageFilter string
}

// NewScenarioAnalyzer creates a new scenario analyzer.
func NewScenarioAnalyzer(repoPath string) *ScenarioAnalyzer {
	return &ScenarioAnalyzer{
		repoPath:   repoPath,
		contextDir: filepath.Join(repoPath, contextDirName),
	}
}

// SetPackageFilter restricts scenario analysis to matching packages.
func (sa *ScenarioAnalyzer) SetPackageFilter(filter string) {
	sa.packageFilter = filter
}

// AnalyzeScenarios generates all test scenarios that should exist.
func (sa *ScenarioAnalyzer) AnalyzeScenarios(ctx context.Context, repoContext *models.RepoContext) (*models.ScenarioAnalysis, error) {
	var allScenarios []models.TestScenario
	scenarioID := 1

	for _, function := range repoContext.AllFunctions {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if strings.HasPrefix(function.Name, "Test") || strings.HasPrefix(function.Name, "Benchmark") {
			continue
		}

		if sa.packageFilter != "" && !strings.Contains(function.Package, sa.packageFilter) {
			continue
		}

		scenarios := generateScenariosForFunction(function, &scenarioID)
		allScenarios = append(allScenarios, scenarios...)
	}

	analysis := &models.ScenarioAnalysis{
		Repository:     sa.repoPath,
		AnalyzedAt:     time.Now(),
		TotalScenarios: len(allScenarios),
		Scenarios:      allScenarios,
	}

	if err := sa.saveAnalysis(analysis); err != nil {
		return nil, fmt.Errorf("failed to save analysis: %w", err)
	}

	return analysis, nil
}

// LoadAnalysis loads previously saved scenario analysis.
func (sa *ScenarioAnalyzer) LoadAnalysis(_ context.Context) (*models.ScenarioAnalysis, error) {
	data, err := os.ReadFile(filepath.Join(sa.contextDir, "scenarios.json"))
	if err != nil {
		return nil, err
	}

	var analysis models.ScenarioAnalysis
	if err := json.Unmarshal(data, &analysis); err != nil {
		return nil, err
	}
	return &analysis, nil
}

func generateScenariosForFunction(fn models.FunctionDetail, scenarioID *int) []models.TestScenario {
	var scenarios []models.TestScenario

	// Happy path: always needed for exported functions, or unexported with complexity > 1
	if fn.Exported || fn.ComplexityScore > 1 {
		scenarios = append(scenarios, models.TestScenario{
			ID:           fmt.Sprintf("S%03d", *scenarioID),
			FunctionID:   fn.ID,
			FunctionName: fn.Name,
			Package:      fn.Package,
			ScenarioType: "happy_path",
			Description:  fmt.Sprintf("Test %s with valid inputs and verify correct output", fn.Name),
			TestName:     fmt.Sprintf("Test%s", fn.Name),
		})
		*scenarioID++
	}

	// Error case: only if the function signature explicitly returns error
	if returnsError(fn.Signature) {
		scenarios = append(scenarios, models.TestScenario{
			ID:           fmt.Sprintf("S%03d", *scenarioID),
			FunctionID:   fn.ID,
			FunctionName: fn.Name,
			Package:      fn.Package,
			ScenarioType: "error_case",
			Description:  fmt.Sprintf("Test %s with invalid inputs that trigger error returns", fn.Name),
			TestName:     fmt.Sprintf("Test%s_Error", fn.Name),
		})
		*scenarioID++
	}

	// Edge case: only for exported functions that are genuinely complex
	if fn.Exported && fn.ComplexityScore > 5 && fn.LineCount > 10 {
		scenarios = append(scenarios, models.TestScenario{
			ID:           fmt.Sprintf("S%03d", *scenarioID),
			FunctionID:   fn.ID,
			FunctionName: fn.Name,
			Package:      fn.Package,
			ScenarioType: "edge_case",
			Description:  fmt.Sprintf("Test %s with edge cases (nil, empty, zero, max values)", fn.Name),
			TestName:     fmt.Sprintf("Test%s_EdgeCases", fn.Name),
		})
		*scenarioID++
	}

	// Boundary: only for exported functions with numeric/collection params AND non-trivial logic
	if fn.Exported && fn.ComplexityScore > 2 && hasNumericOrCollectionParams(fn.Signature) {
		scenarios = append(scenarios, models.TestScenario{
			ID:           fmt.Sprintf("S%03d", *scenarioID),
			FunctionID:   fn.ID,
			FunctionName: fn.Name,
			Package:      fn.Package,
			ScenarioType: "boundary",
			Description:  fmt.Sprintf("Test %s at boundary values (zero, negative, overflow, empty collections)", fn.Name),
			TestName:     fmt.Sprintf("Test%s_Boundary", fn.Name),
		})
		*scenarioID++
	}

	return scenarios
}

// returnsError checks if the function signature has an error return type.
func returnsError(signature string) bool {
	// Look for "error" in the return portion (after the last ")")
	idx := strings.LastIndex(signature, ")")
	if idx == -1 {
		return false
	}
	returnPart := signature[idx:]
	return strings.Contains(returnPart, "error")
}

func hasNumericOrCollectionParams(signature string) bool {
	// Extract the parameters portion between first ( and matching )
	start := strings.Index(signature, "(")
	if start == -1 {
		return false
	}
	// Find the closing paren for params (not return type)
	depth := 0
	end := -1
	for i := start; i < len(signature); i++ {
		if signature[i] == '(' {
			depth++
		} else if signature[i] == ')' {
			depth--
			if depth == 0 {
				end = i
				break
			}
		}
	}
	if end == -1 {
		return false
	}
	params := strings.ToLower(signature[start:end])
	for _, keyword := range []string{"int", "float", "uint", "byte", "[]", "map["} {
		if strings.Contains(params, keyword) {
			return true
		}
	}
	return false
}

func (sa *ScenarioAnalyzer) saveAnalysis(analysis *models.ScenarioAnalysis) error {
	if err := os.MkdirAll(sa.contextDir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(sa.contextDir, "scenarios.json"), data, 0644)
}

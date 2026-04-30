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

		if strings.HasPrefix(function.Name, "Test") {
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

	scenarios = append(scenarios, models.TestScenario{
		ID:           fmt.Sprintf("S%03d", *scenarioID),
		FunctionID:   fn.ID,
		FunctionName: fn.Name,
		Package:      fn.Package,
		ScenarioType: "happy_path",
		Description:  fmt.Sprintf("Test %s with valid inputs", fn.Name),
		TestName:     fmt.Sprintf("Test%s", fn.Name),
	})
	*scenarioID++

	if shouldHaveErrorCase(fn) {
		scenarios = append(scenarios, models.TestScenario{
			ID:           fmt.Sprintf("S%03d", *scenarioID),
			FunctionID:   fn.ID,
			FunctionName: fn.Name,
			Package:      fn.Package,
			ScenarioType: "error_case",
			Description:  fmt.Sprintf("Test %s error handling", fn.Name),
			TestName:     fmt.Sprintf("Test%s_Error", fn.Name),
		})
		*scenarioID++
	}

	if fn.Exported && fn.ComplexityScore > 5 {
		scenarios = append(scenarios, models.TestScenario{
			ID:           fmt.Sprintf("S%03d", *scenarioID),
			FunctionID:   fn.ID,
			FunctionName: fn.Name,
			Package:      fn.Package,
			ScenarioType: "edge_case",
			Description:  fmt.Sprintf("Test %s with edge cases (nil, empty, large values)", fn.Name),
			TestName:     fmt.Sprintf("Test%s_EdgeCases", fn.Name),
		})
		*scenarioID++
	}

	if shouldHaveBoundaryCase(fn) {
		scenarios = append(scenarios, models.TestScenario{
			ID:           fmt.Sprintf("S%03d", *scenarioID),
			FunctionID:   fn.ID,
			FunctionName: fn.Name,
			Package:      fn.Package,
			ScenarioType: "boundary",
			Description:  fmt.Sprintf("Test %s boundary conditions", fn.Name),
			TestName:     fmt.Sprintf("Test%s_Boundary", fn.Name),
		})
		*scenarioID++
	}

	return scenarios
}

func shouldHaveErrorCase(fn models.FunctionDetail) bool {
	if strings.Contains(fn.Signature, "error") {
		return true
	}
	if fn.ReceiverType != "" && strings.Contains(fn.ReceiverType, "*") {
		return true
	}
	return false
}

func shouldHaveBoundaryCase(fn models.FunctionDetail) bool {
	sig := strings.ToLower(fn.Signature)
	for _, keyword := range []string{"int", "int64", "int32", "float", "uint", "[]", "map["} {
		if strings.Contains(sig, keyword) {
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

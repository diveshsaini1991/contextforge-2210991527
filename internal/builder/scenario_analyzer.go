package builder

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/divesh/contextforge/internal/models"
)

// ScenarioAnalyzer generates test scenarios for functions
type ScenarioAnalyzer struct {
	repoPath   string
	contextDir string
}

// NewScenarioAnalyzer creates a new scenario analyzer
func NewScenarioAnalyzer(repoPath string) *ScenarioAnalyzer {
	return &ScenarioAnalyzer{
		repoPath:   repoPath,
		contextDir: filepath.Join(repoPath, contextDirName),
	}
}

// AnalyzeScenarios generates all test scenarios that should exist
func (sa *ScenarioAnalyzer) AnalyzeScenarios(context *models.RepoContext) (*models.ScenarioAnalysis, error) {
	var allScenarios []models.TestScenario
	scenarioID := 1

	for _, function := range context.AllFunctions {
		// Skip test functions themselves
		if strings.HasPrefix(function.Name, "Test") {
			continue
		}

		// Generate scenarios based on function characteristics
		scenarios := sa.generateScenariosForFunction(function, &scenarioID)
		allScenarios = append(allScenarios, scenarios...)
	}

	analysis := &models.ScenarioAnalysis{
		Repository:     sa.repoPath,
		AnalyzedAt:     time.Now(),
		TotalScenarios: len(allScenarios),
		Scenarios:      allScenarios,
	}

	// Save to file
	if err := sa.saveAnalysis(analysis); err != nil {
		return nil, fmt.Errorf("failed to save analysis: %w", err)
	}

	return analysis, nil
}

// generateScenariosForFunction generates test scenarios for a single function
func (sa *ScenarioAnalyzer) generateScenariosForFunction(fn models.FunctionDetail, scenarioID *int) []models.TestScenario {
	var scenarios []models.TestScenario

	// 1. Happy path scenario (always needed)
	scenarios = append(scenarios, models.TestScenario{
		ID:           fmt.Sprintf("S%03d", *scenarioID),
		FunctionID:   fn.ID,
		FunctionName: fn.Name,
		Package:      fn.Package,
		ScenarioType: "happy_path",
		Description:  fmt.Sprintf("Test %s with valid inputs", fn.Name),
		TestName:     fmt.Sprintf("Test%s", fn.Name),
		Exists:       false,
	})
	*scenarioID++

	// 2. Error case scenario (if function returns error or has pointer receivers)
	if sa.shouldHaveErrorCase(fn) {
		scenarios = append(scenarios, models.TestScenario{
			ID:           fmt.Sprintf("S%03d", *scenarioID),
			FunctionID:   fn.ID,
			FunctionName: fn.Name,
			Package:      fn.Package,
			ScenarioType: "error_case",
			Description:  fmt.Sprintf("Test %s error handling", fn.Name),
			TestName:     fmt.Sprintf("Test%s_Error", fn.Name),
			Exists:       false,
		})
		*scenarioID++
	}

	// 3. Edge case scenarios (for exported functions with complexity > 5)
	if fn.Exported && fn.ComplexityScore > 5 {
		scenarios = append(scenarios, models.TestScenario{
			ID:           fmt.Sprintf("S%03d", *scenarioID),
			FunctionID:   fn.ID,
			FunctionName: fn.Name,
			Package:      fn.Package,
			ScenarioType: "edge_case",
			Description:  fmt.Sprintf("Test %s with edge cases (nil, empty, large values)", fn.Name),
			TestName:     fmt.Sprintf("Test%s_EdgeCases", fn.Name),
			Exists:       false,
		})
		*scenarioID++
	}

	// 4. Boundary scenarios (for functions with numeric parameters or slices)
	if sa.shouldHaveBoundaryCase(fn) {
		scenarios = append(scenarios, models.TestScenario{
			ID:           fmt.Sprintf("S%03d", *scenarioID),
			FunctionID:   fn.ID,
			FunctionName: fn.Name,
			Package:      fn.Package,
			ScenarioType: "boundary",
			Description:  fmt.Sprintf("Test %s boundary conditions", fn.Name),
			TestName:     fmt.Sprintf("Test%s_Boundary", fn.Name),
			Exists:       false,
		})
		*scenarioID++
	}

	return scenarios
}

// shouldHaveErrorCase determines if function should have error test cases
func (sa *ScenarioAnalyzer) shouldHaveErrorCase(fn models.FunctionDetail) bool {
	// Check if signature contains error return
	if strings.Contains(fn.Signature, "error") {
		return true
	}
	// Pointer receivers might have error cases
	if fn.ReceiverType != "" && strings.Contains(fn.ReceiverType, "*") {
		return true
	}
	return false
}

// shouldHaveBoundaryCase determines if function should have boundary test cases
func (sa *ScenarioAnalyzer) shouldHaveBoundaryCase(fn models.FunctionDetail) bool {
	// Check for numeric types or slices in signature
	sig := strings.ToLower(fn.Signature)
	numerics := []string{"int", "int64", "int32", "float", "uint"}
	containers := []string{"[]", "map["}

	for _, num := range numerics {
		if strings.Contains(sig, num) {
			return true
		}
	}
	for _, cont := range containers {
		if strings.Contains(sig, cont) {
			return true
		}
	}
	return false
}

// saveAnalysis saves the scenario analysis to a JSON file
func (sa *ScenarioAnalyzer) saveAnalysis(analysis *models.ScenarioAnalysis) error {
	if err := os.MkdirAll(sa.contextDir, 0755); err != nil {
		return err
	}

	scenarioPath := filepath.Join(sa.contextDir, "scenarios.json")
	data, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(scenarioPath, data, 0644)
}

// LoadAnalysis loads previously saved scenario analysis
func (sa *ScenarioAnalyzer) LoadAnalysis() (*models.ScenarioAnalysis, error) {
	scenarioPath := filepath.Join(sa.contextDir, "scenarios.json")
	data, err := os.ReadFile(scenarioPath)
	if err != nil {
		return nil, err
	}

	var analysis models.ScenarioAnalysis
	if err := json.Unmarshal(data, &analysis); err != nil {
		return nil, err
	}

	return &analysis, nil
}

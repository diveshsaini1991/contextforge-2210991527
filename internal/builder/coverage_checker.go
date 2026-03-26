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

// CoverageChecker checks which test scenarios are covered by existing tests
type CoverageChecker struct {
	repoPath   string
	contextDir string
}

// NewCoverageChecker creates a new coverage checker
func NewCoverageChecker(repoPath string) *CoverageChecker {
	return &CoverageChecker{
		repoPath:   repoPath,
		contextDir: filepath.Join(repoPath, contextDirName),
	}
}

// CheckCoverage compares scenarios against existing tests
func (cc *CoverageChecker) CheckCoverage(context *models.RepoContext, scenarios *models.ScenarioAnalysis) (*models.TestCoverageReport, error) {
	// Build a map of existing tests for quick lookup
	existingTests := cc.buildTestMap(context)

	// Check each scenario
	covered := 0
	missing := []models.TestScenario{}
	packageCoverage := make(map[string]*models.PackageReport)

	for i := range scenarios.Scenarios {
		scenario := &scenarios.Scenarios[i]

		// Initialize package report if not exists
		if _, exists := packageCoverage[scenario.Package]; !exists {
			packageCoverage[scenario.Package] = &models.PackageReport{
				Package:          scenario.Package,
				TotalScenarios:   0,
				CoveredScenarios: 0,
			}
		}
		packageCoverage[scenario.Package].TotalScenarios++

		// Check if test exists
		if cc.scenarioIsCovered(scenario, existingTests) {
			scenario.Exists = true
			covered++
			packageCoverage[scenario.Package].CoveredScenarios++
		} else {
			missing = append(missing, *scenario)
		}
	}

	// Calculate package coverage percentages
	var pkgReports []models.PackageReport
	for _, report := range packageCoverage {
		if report.TotalScenarios > 0 {
			report.CoveragePercent = float64(report.CoveredScenarios) / float64(report.TotalScenarios) * 100
		}
		pkgReports = append(pkgReports, *report)
	}

	// Overall coverage
	coveragePercent := 0.0
	if scenarios.TotalScenarios > 0 {
		coveragePercent = float64(covered) / float64(scenarios.TotalScenarios) * 100
	}

	// Generate summary
	summary := cc.generateSummary(covered, len(missing), scenarios.TotalScenarios, coveragePercent)

	report := &models.TestCoverageReport{
		Repository:       cc.repoPath,
		GeneratedAt:      time.Now(),
		TotalScenarios:   scenarios.TotalScenarios,
		CoveredScenarios: covered,
		MissingScenarios: len(missing),
		CoveragePercent:  coveragePercent,
		ByPackage:        pkgReports,
		MissingTests:     missing,
		Summary:          summary,
	}

	// Save report
	if err := cc.saveReport(report); err != nil {
		return nil, fmt.Errorf("failed to save report: %w", err)
	}

	return report, nil
}

// buildTestMap creates a map of function -> test names
func (cc *CoverageChecker) buildTestMap(context *models.RepoContext) map[string][]string {
	testMap := make(map[string][]string)

	for _, fn := range context.AllFunctions {
		if len(fn.ExistingTests) > 0 {
			testMap[fn.ID] = fn.ExistingTests
		}
	}

	return testMap
}

// scenarioIsCovered checks if a scenario is covered by existing tests
func (cc *CoverageChecker) scenarioIsCovered(scenario *models.TestScenario, existingTests map[string][]string) bool {
	tests, exists := existingTests[scenario.FunctionID]
	if !exists {
		return false
	}

	// Check if any existing test matches the scenario test name pattern
	for _, testName := range tests {
		if cc.testMatchesScenario(testName, scenario) {
			scenario.FoundIn = testName
			return true
		}
	}

	return false
}

// testMatchesScenario checks if a test name matches a scenario
func (cc *CoverageChecker) testMatchesScenario(testName string, scenario *models.TestScenario) bool {
	// Exact match
	if testName == scenario.TestName {
		return true
	}

	// Partial match based on scenario type
	testLower := strings.ToLower(testName)
	funcLower := strings.ToLower(scenario.FunctionName)

	switch scenario.ScenarioType {
	case "happy_path":
		// Basic test for the function (Test<FuncName>)
		return strings.Contains(testLower, funcLower) &&
			!strings.Contains(testLower, "error") &&
			!strings.Contains(testLower, "edge") &&
			!strings.Contains(testLower, "boundary")
	case "error_case":
		return strings.Contains(testLower, funcLower) &&
			(strings.Contains(testLower, "error") || strings.Contains(testLower, "fail"))
	case "edge_case":
		return strings.Contains(testLower, funcLower) &&
			(strings.Contains(testLower, "edge") || strings.Contains(testLower, "nil") || strings.Contains(testLower, "empty"))
	case "boundary":
		return strings.Contains(testLower, funcLower) &&
			(strings.Contains(testLower, "boundary") || strings.Contains(testLower, "limit"))
	}

	return false
}

// generateSummary creates a human-readable summary
func (cc *CoverageChecker) generateSummary(covered, missing, total int, percent float64) string {
	return fmt.Sprintf(
		"Test Coverage Analysis:\n"+
			"- Total Scenarios: %d\n"+
			"- Covered: %d (%.1f%%)\n"+
			"- Missing: %d\n"+
			"\nNext Steps:\n"+
			"- Use 'generate_test_stubs' to create missing test functions\n"+
			"- Implement test logic for each stub",
		total, covered, percent, missing,
	)
}

// saveReport saves the coverage report to a JSON file
func (cc *CoverageChecker) saveReport(report *models.TestCoverageReport) error {
	if err := os.MkdirAll(cc.contextDir, 0755); err != nil {
		return err
	}

	reportPath := filepath.Join(cc.contextDir, "coverage_report.json")
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(reportPath, data, 0644)
}

// LoadReport loads previously saved coverage report
func (cc *CoverageChecker) LoadReport() (*models.TestCoverageReport, error) {
	reportPath := filepath.Join(cc.contextDir, "coverage_report.json")
	data, err := os.ReadFile(reportPath)
	if err != nil {
		return nil, err
	}

	var report models.TestCoverageReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, err
	}

	return &report, nil
}

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

// CoverageChecker checks which test scenarios are covered by existing tests.
type CoverageChecker struct {
	repoPath   string
	contextDir string
}

// NewCoverageChecker creates a new coverage checker.
func NewCoverageChecker(repoPath string) *CoverageChecker {
	return &CoverageChecker{
		repoPath:   repoPath,
		contextDir: filepath.Join(repoPath, contextDirName),
	}
}

// CheckCoverage compares scenarios against existing tests and produces a gap report.
func (cc *CoverageChecker) CheckCoverage(ctx context.Context, repoContext *models.RepoContext, scenarios *models.ScenarioAnalysis) (*models.TestCoverageReport, error) {
	existingTests := buildTestMap(repoContext)

	covered := 0
	var missing []models.TestScenario
	packageCoverage := make(map[string]*models.PackageReport)

	for i := range scenarios.Scenarios {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		scenario := &scenarios.Scenarios[i]

		if _, exists := packageCoverage[scenario.Package]; !exists {
			packageCoverage[scenario.Package] = &models.PackageReport{
				Package: scenario.Package,
			}
		}
		packageCoverage[scenario.Package].TotalScenarios++

		if scenarioIsCovered(scenario, existingTests) {
			scenario.Exists = true
			covered++
			packageCoverage[scenario.Package].CoveredScenarios++
		} else {
			missing = append(missing, *scenario)
		}
	}

	var pkgReports []models.PackageReport
	for _, report := range packageCoverage {
		if report.TotalScenarios > 0 {
			report.CoveragePercent = float64(report.CoveredScenarios) / float64(report.TotalScenarios) * 100
		}
		pkgReports = append(pkgReports, *report)
	}

	coveragePercent := 0.0
	if scenarios.TotalScenarios > 0 {
		coveragePercent = float64(covered) / float64(scenarios.TotalScenarios) * 100
	}

	report := &models.TestCoverageReport{
		Repository:       cc.repoPath,
		GeneratedAt:      time.Now(),
		TotalScenarios:   scenarios.TotalScenarios,
		CoveredScenarios: covered,
		MissingScenarios: len(missing),
		CoveragePercent:  coveragePercent,
		ByPackage:        pkgReports,
		MissingTests:     missing,
		Summary: fmt.Sprintf(
			"Test Coverage: %d/%d scenarios covered (%.1f%%). Missing: %d. "+
				"Use generate_test_stubs to create stubs for missing tests.",
			covered, scenarios.TotalScenarios, coveragePercent, len(missing),
		),
	}

	if err := cc.saveReport(report); err != nil {
		return nil, fmt.Errorf("failed to save report: %w", err)
	}

	return report, nil
}

// LoadReport loads previously saved coverage report.
func (cc *CoverageChecker) LoadReport(_ context.Context) (*models.TestCoverageReport, error) {
	data, err := os.ReadFile(filepath.Join(cc.contextDir, "coverage_report.json"))
	if err != nil {
		return nil, err
	}

	var report models.TestCoverageReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, err
	}
	return &report, nil
}

func buildTestMap(repoContext *models.RepoContext) map[string][]string {
	testMap := make(map[string][]string)
	for _, fn := range repoContext.AllFunctions {
		if len(fn.ExistingTests) > 0 {
			testMap[fn.ID] = fn.ExistingTests
		}
	}
	return testMap
}

func scenarioIsCovered(scenario *models.TestScenario, existingTests map[string][]string) bool {
	tests, exists := existingTests[scenario.FunctionID]
	if !exists {
		return false
	}

	for _, testName := range tests {
		if testMatchesScenario(testName, scenario) {
			scenario.FoundIn = testName
			return true
		}
	}
	return false
}

func testMatchesScenario(testName string, scenario *models.TestScenario) bool {
	if testName == scenario.TestName {
		return true
	}

	testLower := strings.ToLower(testName)
	funcLower := strings.ToLower(scenario.FunctionName)

	switch scenario.ScenarioType {
	case "happy_path":
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

func (cc *CoverageChecker) saveReport(report *models.TestCoverageReport) error {
	if err := os.MkdirAll(cc.contextDir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(cc.contextDir, "coverage_report.json"), data, 0644)
}

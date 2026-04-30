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

// testEvidence holds what we know about tests for a function by reading the test file content.
type testEvidence struct {
	hasTest         bool
	testNames       []string
	isTableDriven   bool
	tableCaseCount  int
	hasErrorCheck   bool
	hasNilCheck     bool
	hasZeroCheck    bool
	hasBoundaryCase bool
	hasSubtests     bool
}

// CheckCoverage compares scenarios against existing tests and produces a gap report.
func (cc *CoverageChecker) CheckCoverage(ctx context.Context, repoContext *models.RepoContext, scenarios *models.ScenarioAnalysis) (*models.TestCoverageReport, error) {
	// Build evidence by reading actual test file content
	evidence := cc.buildEvidence(repoContext)

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

		if cc.scenarioIsCovered(scenario, evidence) {
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

	scenarioCoverage := 0.0
	if scenarios.TotalScenarios > 0 {
		scenarioCoverage = float64(covered) / float64(scenarios.TotalScenarios) * 100
	}

	// Count function-level coverage
	totalFuncs := 0
	testedFuncs := 0
	var untestedList []string
	for _, fn := range repoContext.AllFunctions {
		if strings.HasPrefix(fn.Name, "Test") || strings.HasPrefix(fn.Name, "Benchmark") {
			continue
		}
		totalFuncs++
		if fn.HasTests {
			testedFuncs++
		} else {
			untestedList = append(untestedList, fmt.Sprintf("%s.%s", fn.Package, fn.Name))
		}
	}
	funcCoverage := 0.0
	if totalFuncs > 0 {
		funcCoverage = float64(testedFuncs) / float64(totalFuncs) * 100
	}

	report := &models.TestCoverageReport{
		Repository:        cc.repoPath,
		GeneratedAt:       time.Now(),
		TotalFunctions:    totalFuncs,
		TestedFunctions:   testedFuncs,
		UntestedFunctions: totalFuncs - testedFuncs,
		FunctionCoverage:  funcCoverage,
		TotalScenarios:    scenarios.TotalScenarios,
		CoveredScenarios:  covered,
		MissingScenarios:  len(missing),
		ScenarioCoverage:  scenarioCoverage,
		ByPackage:         pkgReports,
		UntestedFuncList:  untestedList,
		MissingTests:      missing,
		Summary: fmt.Sprintf(
			"Functions: %d/%d have tests (%.0f%%). "+
				"Scenarios: %d/%d covered (%.0f%%). "+
				"Untested functions: %d. Missing scenarios: %d.",
			testedFuncs, totalFuncs, funcCoverage,
			covered, scenarios.TotalScenarios, scenarioCoverage,
			totalFuncs-testedFuncs, len(missing),
		),
	}

	if err := cc.saveReport(report); err != nil {
		return nil, fmt.Errorf("failed to save report: %w", err)
	}

	return report, nil
}

// buildEvidence reads test files and builds evidence about what each function's tests actually cover.
func (cc *CoverageChecker) buildEvidence(repoContext *models.RepoContext) map[string]*testEvidence {
	evidenceMap := make(map[string]*testEvidence)

	// Read all test file contents by package
	testFileContents := make(map[string]string) // package path -> concatenated test content
	for _, pkg := range repoContext.Packages {
		for _, testFile := range pkg.TestFiles {
			data, err := os.ReadFile(testFile)
			if err != nil {
				continue
			}
			testFileContents[pkg.Path] += "\n" + string(data)
		}
	}

	for _, fn := range repoContext.AllFunctions {
		ev := &testEvidence{}

		if len(fn.ExistingTests) > 0 {
			ev.hasTest = true
			ev.testNames = fn.ExistingTests

			// Analyze the actual test content for this function
			content := testFileContents[cc.findPackagePath(repoContext, fn.Package)]
			if content != "" {
				cc.analyzeTestContent(ev, fn.Name, content)
			}
		}

		evidenceMap[fn.ID] = ev
	}

	return evidenceMap
}

func (cc *CoverageChecker) findPackagePath(repoContext *models.RepoContext, pkgName string) string {
	for _, pkg := range repoContext.Packages {
		if pkg.Name == pkgName {
			return pkg.Path
		}
	}
	return ""
}

// analyzeTestContent reads the test source and determines what scenarios the tests cover.
func (cc *CoverageChecker) analyzeTestContent(ev *testEvidence, funcName string, content string) {
	contentLower := strings.ToLower(content)

	// Find the test function(s) for this function
	for _, testName := range ev.testNames {
		testBlock := cc.extractTestBlock(content, testName)
		if testBlock == "" {
			continue
		}
		blockLower := strings.ToLower(testBlock)

		// Table-driven test detection
		if strings.Contains(testBlock, "[]struct") || strings.Contains(testBlock, "[]struct{") {
			ev.isTableDriven = true
			ev.tableCaseCount = strings.Count(testBlock, "{\"") + strings.Count(testBlock, "{\n")
			if ev.tableCaseCount < 2 {
				ev.tableCaseCount = countTableEntries(testBlock)
			}
		}

		// Subtest detection
		if strings.Contains(testBlock, "t.Run(") {
			ev.hasSubtests = true
		}

		// Error checking detection
		if strings.Contains(blockLower, "err !=") ||
			strings.Contains(blockLower, "error") ||
			strings.Contains(blockLower, "wanterr") ||
			strings.Contains(blockLower, "expecterr") ||
			strings.Contains(blockLower, "shoulderr") {
			ev.hasErrorCheck = true
		}

		// Nil/zero/empty detection
		if strings.Contains(blockLower, "nil") ||
			strings.Contains(blockLower, "empty") {
			ev.hasNilCheck = true
		}

		// Zero/negative/boundary detection
		if strings.Contains(blockLower, "zero") ||
			strings.Contains(blockLower, "negative") ||
			strings.Contains(blockLower, ", 0,") ||
			strings.Contains(blockLower, ", 0}") ||
			strings.Contains(blockLower, ", -") {
			ev.hasZeroCheck = true
		}

		// Boundary value detection in table entries
		if strings.Contains(blockLower, "max") ||
			strings.Contains(blockLower, "min") ||
			strings.Contains(blockLower, "overflow") ||
			strings.Contains(blockLower, "limit") ||
			strings.Contains(blockLower, "boundary") {
			ev.hasBoundaryCase = true
		}
	}

	// Also check package-level for subtests referencing this function
	funcLower := strings.ToLower(funcName)
	if strings.Contains(contentLower, funcLower) {
		if strings.Contains(contentLower, "\"error") || strings.Contains(contentLower, "\"fail") {
			ev.hasErrorCheck = true
		}
	}
}

// extractTestBlock extracts the body of a test function from source.
func (cc *CoverageChecker) extractTestBlock(content, testName string) string {
	marker := "func " + testName + "("
	idx := strings.Index(content, marker)
	if idx == -1 {
		return ""
	}

	// Find the opening brace
	braceIdx := strings.Index(content[idx:], "{")
	if braceIdx == -1 {
		return ""
	}
	start := idx + braceIdx

	// Match braces to find the end
	depth := 0
	for i := start; i < len(content); i++ {
		if content[i] == '{' {
			depth++
		} else if content[i] == '}' {
			depth--
			if depth == 0 {
				return content[start : i+1]
			}
		}
	}
	return content[start:]
}

// countTableEntries counts struct literal entries in a table-driven test.
func countTableEntries(testBlock string) int {
	count := 0
	lines := strings.Split(testBlock, "\n")
	inTable := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "[]struct") {
			inTable = true
			continue
		}
		if inTable {
			if trimmed == "}" || trimmed == "})" {
				inTable = false
				continue
			}
			if strings.HasPrefix(trimmed, "{") {
				count++
			}
		}
	}
	if count == 0 {
		count = 1
	}
	return count
}

// scenarioIsCovered decides if a scenario is covered based on test evidence.
func (cc *CoverageChecker) scenarioIsCovered(scenario *models.TestScenario, evidence map[string]*testEvidence) bool {
	ev, exists := evidence[scenario.FunctionID]
	if !exists || !ev.hasTest {
		return false
	}

	switch scenario.ScenarioType {
	case "happy_path":
		// If tests exist at all, happy path is covered
		scenario.FoundIn = strings.Join(ev.testNames, ", ")
		return true

	case "error_case":
		if ev.hasErrorCheck {
			scenario.FoundIn = strings.Join(ev.testNames, ", ") + " (error assertions found)"
			return true
		}
		// Table-driven with 3+ cases likely covers errors
		if ev.isTableDriven && ev.tableCaseCount >= 3 && ev.hasErrorCheck {
			scenario.FoundIn = strings.Join(ev.testNames, ", ") + " (table-driven with error cases)"
			return true
		}
		return false

	case "edge_case":
		if ev.hasNilCheck || ev.hasZeroCheck {
			scenario.FoundIn = strings.Join(ev.testNames, ", ") + " (nil/zero checks found)"
			return true
		}
		// Table-driven with 4+ cases likely covers edge cases
		if ev.isTableDriven && ev.tableCaseCount >= 4 {
			scenario.FoundIn = strings.Join(ev.testNames, ", ") + " (table-driven with multiple cases)"
			return true
		}
		return false

	case "boundary":
		if ev.hasBoundaryCase {
			scenario.FoundIn = strings.Join(ev.testNames, ", ") + " (boundary values found)"
			return true
		}
		// Table-driven test with zero/negative values covers basic boundaries
		if ev.isTableDriven && ev.hasZeroCheck {
			scenario.FoundIn = strings.Join(ev.testNames, ", ") + " (table-driven with zero/negative values)"
			return true
		}
		return false
	}

	return false
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

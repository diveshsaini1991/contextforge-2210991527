package coverage

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/divesh/contextforge/internal/models"
)

// Analyzer performs coverage analysis on a repository
type Analyzer struct {
	repoPath string
	context  *models.RepositoryContext
}

// NewAnalyzer creates a new coverage analyzer
func NewAnalyzer(repoPath string, context *models.RepositoryContext) *Analyzer {
	return &Analyzer{
		repoPath: repoPath,
		context:  context,
	}
}

// Analyze runs coverage analysis and generates a report
func (a *Analyzer) Analyze() (*models.CoverageReport, error) {
	// Run tests with coverage
	coverageFile := filepath.Join(a.repoPath, "coverage.out")
	if err := a.runCoverage(coverageFile); err != nil {
		return nil, fmt.Errorf("failed to run coverage: %w", err)
	}
	defer os.Remove(coverageFile)

	// Parse coverage profile
	profile, err := ParseCoverageProfile(coverageFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse coverage: %w", err)
	}

	// Build coverage report
	report := a.buildReport(profile)

	return report, nil
}

// runCoverage executes go test with coverage
func (a *Analyzer) runCoverage(outputFile string) error {
	cmd := exec.Command("go", "test", "-cover", "-coverprofile="+outputFile, "./...")
	cmd.Dir = a.repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// buildReport builds a coverage report from the profile and context
func (a *Analyzer) buildReport(profile *CoverageProfile) *models.CoverageReport {
	report := &models.CoverageReport{
		AnalyzedAt:        time.Now(),
		CoverageByPackage: []models.PackageCoverage{},
	}

	// Process each package
	for _, pkg := range a.context.Packages {
		pkgCov := a.analyzePackage(pkg, profile)
		report.CoverageByPackage = append(report.CoverageByPackage, pkgCov)

		report.TotalUncovered += len(pkgCov.UncoveredFunctions)
		report.TotalPartiallyCovered += len(pkgCov.PartiallyCovered)
		report.TotalFullyCovered += len(pkgCov.FullyCovered)
	}

	// Calculate overall coverage
	totalFuncs := report.TotalUncovered + report.TotalPartiallyCovered + report.TotalFullyCovered
	if totalFuncs > 0 {
		report.OverallCoverage = float64(report.TotalFullyCovered) / float64(totalFuncs) * 100.0
	}

	return report
}

// analyzePackage analyzes coverage for a single package
func (a *Analyzer) analyzePackage(pkg models.PackageInfo, profile *CoverageProfile) models.PackageCoverage {
	pkgCov := models.PackageCoverage{
		Package:            pkg.Name,
		UncoveredFunctions: []models.FunctionCoverage{},
		PartiallyCovered:   []models.FunctionCoverage{},
		FullyCovered:       []models.FunctionCoverage{},
	}

	totalCoverage := 0.0
	functionCount := 0

	for _, fn := range pkg.Functions {
		funcCov := a.analyzeFunctionCoverage(fn, profile)

		totalCoverage += funcCov.CurrentCoverage
		functionCount++

		switch funcCov.Status {
		case models.CoverageUncovered:
			pkgCov.UncoveredFunctions = append(pkgCov.UncoveredFunctions, funcCov)
		case models.CoveragePartial:
			pkgCov.PartiallyCovered = append(pkgCov.PartiallyCovered, funcCov)
		case models.CoverageFull:
			pkgCov.FullyCovered = append(pkgCov.FullyCovered, funcCov)
		}
	}

	if functionCount > 0 {
		pkgCov.CoveragePercent = totalCoverage / float64(functionCount)
	}

	return pkgCov
}

// analyzeFunctionCoverage analyzes coverage for a single function
func (a *Analyzer) analyzeFunctionCoverage(fn models.FunctionInfo, profile *CoverageProfile) models.FunctionCoverage {
	funcCov := models.FunctionCoverage{
		FunctionInfo:    fn,
		CurrentCoverage: 0.0,
		Status:          models.CoverageUncovered,
	}

	// Find coverage data for this file
	var fileCov *FileCoverage
	for path, fc := range profile.FileCoverage {
		if strings.HasSuffix(path, fn.File) || strings.HasSuffix(fn.File, path) {
			fileCov = fc
			break
		}
	}

	if fileCov == nil {
		return funcCov
	}

	// Calculate coverage for this function's line range
	lineCoverage := fileCov.GetLineCoverage()
	coverage := CalculateFunctionCoverage(fn.StartLine, fn.EndLine, lineCoverage)

	funcCov.CurrentCoverage = coverage

	// Determine status
	if coverage == 0.0 {
		funcCov.Status = models.CoverageUncovered
	} else if coverage >= 100.0 {
		funcCov.Status = models.CoverageFull
	} else {
		funcCov.Status = models.CoveragePartial
	}

	return funcCov
}

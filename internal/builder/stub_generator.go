package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/divesh/contextforge/internal/models"
)

// StubGenerator generates test stub files
type StubGenerator struct {
	repoPath string
}

// NewStubGenerator creates a new stub generator
func NewStubGenerator(repoPath string) *StubGenerator {
	return &StubGenerator{
		repoPath: repoPath,
	}
}

// GenerateStubs creates test files with stub functions for missing tests
func (sg *StubGenerator) GenerateStubs(context *models.RepoContext, report *models.TestCoverageReport) ([]string, error) {
	// Group missing scenarios by package
	byPackage := sg.groupByPackage(report.MissingTests)

	var createdFiles []string

	for pkgName, scenarios := range byPackage {
		// Find the package info
		var pkgInfo *models.PackageDetail
		for i := range context.Packages {
			if context.Packages[i].Name == pkgName {
				pkgInfo = &context.Packages[i]
				break
			}
		}

		if pkgInfo == nil {
			continue
		}

		// Determine test file path
		testFilePath := sg.getTestFilePath(pkgInfo)

		// Check if test file already exists
		fileExists := false
		var existingContent string
		if data, err := os.ReadFile(testFilePath); err == nil {
			fileExists = true
			existingContent = string(data)
		}

		// Generate test stubs
		stubs := sg.generateTestStubs(pkgName, scenarios, existingContent)

		// Write or append to test file
		if fileExists {
			// Append new stubs
			if err := sg.appendStubs(testFilePath, stubs, existingContent); err != nil {
				return createdFiles, err
			}
		} else {
			// Create new test file
			if err := sg.createTestFile(testFilePath, pkgName, stubs); err != nil {
				return createdFiles, err
			}
		}

		createdFiles = append(createdFiles, testFilePath)
	}

	return createdFiles, nil
}

// groupByPackage groups scenarios by package name
func (sg *StubGenerator) groupByPackage(scenarios []models.TestScenario) map[string][]models.TestScenario {
	byPackage := make(map[string][]models.TestScenario)

	for _, scenario := range scenarios {
		byPackage[scenario.Package] = append(byPackage[scenario.Package], scenario)
	}

	return byPackage
}

// getTestFilePath determines the test file path for a package
func (sg *StubGenerator) getTestFilePath(pkg *models.PackageDetail) string {
	// Use the package path to determine test file location
	packageDir := filepath.Join(sg.repoPath, pkg.Path)

	// If test files already exist, use the first one
	if len(pkg.TestFiles) > 0 {
		return pkg.TestFiles[0]
	}

	// Otherwise, create a new test file based on package name
	return filepath.Join(packageDir, fmt.Sprintf("%s_test.go", pkg.Name))
}

// generateTestStubs generates stub function code
func (sg *StubGenerator) generateTestStubs(pkgName string, scenarios []models.TestScenario, existingContent string) string {
	var stubs strings.Builder

	for _, scenario := range scenarios {
		// Skip if stub already exists
		if strings.Contains(existingContent, fmt.Sprintf("func %s(", scenario.TestName)) {
			continue
		}

		stubs.WriteString(fmt.Sprintf("\n// %s\n", scenario.Description))
		stubs.WriteString(fmt.Sprintf("func %s(t *testing.T) {\n", scenario.TestName))
		stubs.WriteString("\t// TODO: Implement test logic\n")
		stubs.WriteString(fmt.Sprintf("\tt.Log(\"Test not yet built: %s\")\n", scenario.Description))
		stubs.WriteString("\tt.Skip(\"Test stub - implementation pending\")\n")
		stubs.WriteString("}\n")
	}

	return stubs.String()
}

// createTestFile creates a new test file with stubs
func (sg *StubGenerator) createTestFile(filePath, pkgName string, stubs string) error {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var content strings.Builder
	content.WriteString(fmt.Sprintf("package %s\n\n", pkgName))
	content.WriteString("import (\n")
	content.WriteString("\t\"testing\"\n")
	content.WriteString(")\n")
	content.WriteString(stubs)

	return os.WriteFile(filePath, []byte(content.String()), 0644)
}

// appendStubs appends new stubs to existing test file
func (sg *StubGenerator) appendStubs(filePath, stubs, existingContent string) error {
	if stubs == "" {
		return nil // No new stubs to add
	}

	// Append new stubs at the end
	newContent := existingContent
	if !strings.HasSuffix(existingContent, "\n") {
		newContent += "\n"
	}
	newContent += stubs

	return os.WriteFile(filePath, []byte(newContent), 0644)
}

// GenerateStubsForScenarios is a convenience method that takes scenario IDs
func (sg *StubGenerator) GenerateStubsForScenarios(context *models.RepoContext, scenarios []models.TestScenario) ([]string, error) {
	// Group scenarios by package
	byPackage := sg.groupByPackage(scenarios)

	var createdFiles []string

	for pkgName, scenarioList := range byPackage {
		// Find package info
		var pkgInfo *models.PackageDetail
		for i := range context.Packages {
			if context.Packages[i].Name == pkgName {
				pkgInfo = &context.Packages[i]
				break
			}
		}

		if pkgInfo == nil {
			continue
		}

		// Determine test file path
		testFilePath := sg.getTestFilePath(pkgInfo)

		// Check if file exists
		fileExists := false
		var existingContent string
		if data, err := os.ReadFile(testFilePath); err == nil {
			fileExists = true
			existingContent = string(data)
		}

		// Generate stubs
		stubs := sg.generateTestStubs(pkgName, scenarioList, existingContent)

		// Write file
		if fileExists {
			if err := sg.appendStubs(testFilePath, stubs, existingContent); err != nil {
				return createdFiles, err
			}
		} else {
			if err := sg.createTestFile(testFilePath, pkgName, stubs); err != nil {
				return createdFiles, err
			}
		}

		createdFiles = append(createdFiles, testFilePath)
	}

	return createdFiles, nil
}

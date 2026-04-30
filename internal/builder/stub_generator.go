package builder

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/divesh/contextforge/internal/models"
)

// StubGenerator generates test stub files for missing tests.
type StubGenerator struct {
	repoPath string
}

// NewStubGenerator creates a new stub generator.
func NewStubGenerator(repoPath string) *StubGenerator {
	return &StubGenerator{repoPath: repoPath}
}

// GenerateStubs creates test files with stub functions for missing tests.
func (sg *StubGenerator) GenerateStubs(ctx context.Context, repoContext *models.RepoContext, report *models.TestCoverageReport) ([]string, error) {
	byPackage := groupByPackage(report.MissingTests)
	var createdFiles []string

	for pkgName, scenarios := range byPackage {
		if ctx.Err() != nil {
			return createdFiles, ctx.Err()
		}

		pkgInfo := findPackage(repoContext, pkgName)
		if pkgInfo == nil {
			continue
		}

		testFilePath := getTestFilePath(sg.repoPath, pkgInfo)

		var existingContent string
		fileExists := false
		if data, err := os.ReadFile(testFilePath); err == nil {
			fileExists = true
			existingContent = string(data)
		}

		stubs := generateTestStubs(scenarios, existingContent)
		if stubs == "" {
			continue
		}

		if fileExists {
			if err := appendStubs(testFilePath, stubs, existingContent); err != nil {
				return createdFiles, fmt.Errorf("failed to append stubs to %s: %w", testFilePath, err)
			}
		} else {
			if err := createTestFile(testFilePath, pkgName, stubs); err != nil {
				return createdFiles, fmt.Errorf("failed to create test file %s: %w", testFilePath, err)
			}
		}

		createdFiles = append(createdFiles, testFilePath)
	}

	return createdFiles, nil
}

func findPackage(repoContext *models.RepoContext, pkgName string) *models.PackageDetail {
	for i := range repoContext.Packages {
		if repoContext.Packages[i].Name == pkgName {
			return &repoContext.Packages[i]
		}
	}
	return nil
}

func groupByPackage(scenarios []models.TestScenario) map[string][]models.TestScenario {
	byPackage := make(map[string][]models.TestScenario)
	for _, scenario := range scenarios {
		byPackage[scenario.Package] = append(byPackage[scenario.Package], scenario)
	}
	return byPackage
}

func getTestFilePath(repoPath string, pkg *models.PackageDetail) string {
	if len(pkg.TestFiles) > 0 {
		return pkg.TestFiles[0]
	}
	return filepath.Join(repoPath, pkg.Path, fmt.Sprintf("%s_test.go", pkg.Name))
}

func generateTestStubs(scenarios []models.TestScenario, existingContent string) string {
	var stubs strings.Builder

	for _, scenario := range scenarios {
		if strings.Contains(existingContent, fmt.Sprintf("func %s(", scenario.TestName)) {
			continue
		}

		stubs.WriteString(fmt.Sprintf("\n// %s\n", scenario.Description))
		stubs.WriteString(fmt.Sprintf("func %s(t *testing.T) {\n", scenario.TestName))
		stubs.WriteString(fmt.Sprintf("\tt.Skip(%q)\n", "stub: "+scenario.Description))
		stubs.WriteString("}\n")
	}

	return stubs.String()
}

func createTestFile(filePath, pkgName string, stubs string) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}

	var content strings.Builder
	content.WriteString(fmt.Sprintf("package %s\n\nimport \"testing\"\n", pkgName))
	content.WriteString(stubs)

	return os.WriteFile(filePath, []byte(content.String()), 0644)
}

func appendStubs(filePath, stubs, existingContent string) error {
	newContent := existingContent
	if !strings.HasSuffix(existingContent, "\n") {
		newContent += "\n"
	}
	newContent += stubs
	return os.WriteFile(filePath, []byte(newContent), 0644)
}

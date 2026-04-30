package builder

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/divesh/contextforge/internal/models"
)

// StubGenerator generates test files for missing tests.
type StubGenerator struct {
	repoPath string
}

// NewStubGenerator creates a new stub generator.
func NewStubGenerator(repoPath string) *StubGenerator {
	return &StubGenerator{repoPath: repoPath}
}

// GenerateStubs creates test files with real test implementations for missing tests.
func (sg *StubGenerator) GenerateStubs(ctx context.Context, repoContext *models.RepoContext, report *models.TestCoverageReport) ([]string, error) {
	// Build function lookup from context
	funcLookup := make(map[string]models.FunctionDetail)
	for _, fn := range repoContext.AllFunctions {
		funcLookup[fn.ID] = fn
	}

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

		testCode, extraImports := generateTests(scenarios, existingContent, funcLookup)
		if testCode == "" {
			continue
		}

		if fileExists {
			if err := appendStubs(testFilePath, testCode, existingContent); err != nil {
				return createdFiles, fmt.Errorf("failed to append tests to %s: %w", testFilePath, err)
			}
		} else {
			if err := createTestFile(testFilePath, pkgName, testCode, extraImports); err != nil {
				return createdFiles, fmt.Errorf("failed to create test file %s: %w", testFilePath, err)
			}
		}

		fixImports(testFilePath)
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

func generateTests(scenarios []models.TestScenario, existingContent string, funcLookup map[string]models.FunctionDetail) (string, []string) {
	var sb strings.Builder
	allImports := map[string]bool{}

	for _, scenario := range scenarios {
		if strings.Contains(existingContent, fmt.Sprintf("func %s(", scenario.TestName)) {
			continue
		}

		fn, found := funcLookup[scenario.FunctionID]
		if found && len(fn.Params) > 0 || (found && len(fn.Returns) > 0) {
			code, imports := GenerateTestCode(fn, scenario)
			sb.WriteString(code)
			for _, imp := range imports {
				allImports[imp] = true
			}
		} else {
			sb.WriteString(fmt.Sprintf("\nfunc %s(t *testing.T) {\n", scenario.TestName))
			sb.WriteString(fmt.Sprintf("\tt.Skip(%q)\n", "stub: "+scenario.Description))
			sb.WriteString("}\n")
		}
	}

	var importList []string
	for imp := range allImports {
		importList = append(importList, imp)
	}
	sort.Strings(importList)

	return sb.String(), importList
}

func createTestFile(filePath, pkgName string, testCode string, extraImports []string) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}

	var content strings.Builder
	content.WriteString(fmt.Sprintf("package %s\n\nimport (\n\t\"testing\"\n", pkgName))
	for _, imp := range extraImports {
		content.WriteString(fmt.Sprintf("\t%q\n", imp))
	}
	content.WriteString(")\n")
	content.WriteString(testCode)

	return os.WriteFile(filePath, []byte(content.String()), 0644)
}

func appendStubs(filePath, testCode, existingContent string) error {
	newContent := existingContent
	if !strings.HasSuffix(existingContent, "\n") {
		newContent += "\n"
	}
	newContent += testCode
	return os.WriteFile(filePath, []byte(newContent), 0644)
}

func fixImports(filePath string) {
	_ = exec.Command("goimports", "-w", filePath).Run()
}

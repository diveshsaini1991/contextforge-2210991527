package builder

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/divesh/contextforge/internal/models"
	"github.com/divesh/contextforge/internal/scanner"
)

const contextDirName = ".contextforge"

// ContextBuilder builds and persists repository context
type ContextBuilder struct {
	repoPath   string
	contextDir string
}

// NewContextBuilder creates a new context builder
func NewContextBuilder(repoPath string) *ContextBuilder {
	contextDir := filepath.Join(repoPath, contextDirName)
	return &ContextBuilder{
		repoPath:   repoPath,
		contextDir: contextDir,
	}
}

// BuildContext scans the repository and creates comprehensive context
func (cb *ContextBuilder) BuildContext() (*models.RepoContext, error) {
	// Ensure context directory exists
	if err := os.MkdirAll(cb.contextDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create context directory: %w", err)
	}

	// Scan all Go files (including tests this time)
	packages, err := cb.scanAllFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to scan files: %w", err)
	}

	// Build function details with test information
	allFunctions := []models.FunctionDetail{}
	packageDetails := []models.PackageDetail{}
	totalTests := 0

	for pkgPath, pkgInfo := range packages {
		pkgDetail := models.PackageDetail{
			Name:        pkgInfo.Name,
			Path:        pkgPath,
			Description: fmt.Sprintf("Package %s", pkgInfo.Name),
			Functions:   []models.FunctionDetail{},
			TestFiles:   pkgInfo.TestFiles,
		}

		// Process each function
		for _, fn := range pkgInfo.Functions {
			funcDetail := models.FunctionDetail{
				ID:              fmt.Sprintf("%s.%s", pkgInfo.Name, fn.Name),
				Name:            fn.Name,
				Package:         pkgInfo.Name,
				Signature:       fn.Signature,
				Exported:        fn.Exported,
				FilePath:        fn.File,
				StartLine:       fn.StartLine,
				EndLine:         fn.EndLine,
				LineCount:       fn.LineCount,
				ComplexityScore: fn.ComplexityScore,
				ReceiverType:    fn.ReceiverType,
				HasTests:        false,
				ExistingTests:   []string{},
			}

			// Check if tests exist for this function
			existingTests := cb.findTestsForFunction(fn.Name, pkgInfo.TestFunctions)
			if len(existingTests) > 0 {
				funcDetail.HasTests = true
				funcDetail.ExistingTests = existingTests
				totalTests += len(existingTests)
			}

			pkgDetail.Functions = append(pkgDetail.Functions, funcDetail)
			allFunctions = append(allFunctions, funcDetail)
		}

		packageDetails = append(packageDetails, pkgDetail)
	}

	// Build summary
	summary := models.ContextSummary{
		TotalPackages:  len(packageDetails),
		TotalFunctions: len(allFunctions),
		TotalTests:     totalTests,
		Description:    fmt.Sprintf("Repository with %d packages and %d functions", len(packageDetails), len(allFunctions)),
	}

	context := &models.RepoContext{
		Repository:   cb.repoPath,
		CreatedAt:    time.Now(),
		Summary:      summary,
		Packages:     packageDetails,
		AllFunctions: allFunctions,
	}

	// Save to file
	if err := cb.saveContext(context); err != nil {
		return nil, fmt.Errorf("failed to save context: %w", err)
	}

	return context, nil
}

// scanAllFiles scans both source and test files
func (cb *ContextBuilder) scanAllFiles() (map[string]*packageInfo, error) {
	packages := make(map[string]*packageInfo)

	err := filepath.Walk(cb.repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip context directory
		if strings.Contains(path, contextDirName) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip vendor and hidden directories
		if shouldSkipPath(path) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		pkgPath := getPackagePath(cb.repoPath, path)
		pkgName := getPackageName(path)

		// Initialize package if not exists
		if _, exists := packages[pkgPath]; !exists {
			packages[pkgPath] = &packageInfo{
				Name:          pkgName,
				Functions:     []models.FunctionInfo{},
				TestFunctions: []string{},
				TestFiles:     []string{},
			}
		}

		// Parse the file
		if strings.HasSuffix(path, "_test.go") {
			// Test file - extract test function names
			testFuncs, err := scanner.ParseFile(path)
			if err == nil {
				for _, fn := range testFuncs {
					packages[pkgPath].TestFunctions = append(packages[pkgPath].TestFunctions, fn.Name)
				}
				packages[pkgPath].TestFiles = append(packages[pkgPath].TestFiles, path)
			}
		} else {
			// Source file - extract functions
			functions, err := scanner.ParseFile(path)
			if err == nil {
				packages[pkgPath].Functions = append(packages[pkgPath].Functions, functions...)
			}
		}

		return nil
	})

	return packages, err
}

// findTestsForFunction finds test functions that test a given function
func (cb *ContextBuilder) findTestsForFunction(funcName string, testFunctions []string) []string {
	var matches []string
	searchPatterns := []string{
		fmt.Sprintf("Test%s", funcName),
		fmt.Sprintf("Test_%s", funcName),
		fmt.Sprintf("TestNew%s", funcName),
	}

	for _, testFunc := range testFunctions {
		for _, pattern := range searchPatterns {
			if strings.Contains(testFunc, pattern) {
				matches = append(matches, testFunc)
				break
			}
		}
	}

	return matches
}

// saveContext saves the context to a JSON file
func (cb *ContextBuilder) saveContext(context *models.RepoContext) error {
	contextPath := filepath.Join(cb.contextDir, "context.json")
	data, err := json.MarshalIndent(context, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(contextPath, data, 0644)
}

// LoadContext loads previously saved context
func (cb *ContextBuilder) LoadContext() (*models.RepoContext, error) {
	contextPath := filepath.Join(cb.contextDir, "context.json")
	data, err := os.ReadFile(contextPath)
	if err != nil {
		return nil, err
	}

	var context models.RepoContext
	if err := json.Unmarshal(data, &context); err != nil {
		return nil, err
	}

	return &context, nil
}

// packageInfo holds temporary package data during scanning
type packageInfo struct {
	Name          string
	Functions     []models.FunctionInfo
	TestFunctions []string
	TestFiles     []string
}

// Helper functions
func shouldSkipPath(path string) bool {
	skipDirs := []string{
		"vendor/", ".git/", "node_modules/", ".idea/", ".vscode/",
	}
	for _, skip := range skipDirs {
		if strings.Contains(path, skip) {
			return true
		}
	}
	return false
}

func getPackagePath(repoRoot, filePath string) string {
	dir := filepath.Dir(filePath)
	relPath, err := filepath.Rel(repoRoot, dir)
	if err != nil {
		return dir
	}
	if relPath == "." {
		return "."
	}
	return relPath
}

func getPackageName(filePath string) string {
	dir := filepath.Dir(filePath)
	return filepath.Base(dir)
}

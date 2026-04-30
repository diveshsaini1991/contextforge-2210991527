package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/divesh/contextforge/internal/models"
	"github.com/divesh/contextforge/internal/scanner"
)

const contextDirName = ".contextforge"

// ContextBuilder builds and persists repository context.
type ContextBuilder struct {
	repoPath      string
	contextDir    string
	packageFilter string
}

// NewContextBuilder creates a new context builder for the given repository path.
func NewContextBuilder(repoPath string) *ContextBuilder {
	return &ContextBuilder{
		repoPath:   repoPath,
		contextDir: filepath.Join(repoPath, contextDirName),
	}
}

// SetPackageFilter restricts scanning to packages whose path contains the given substring.
func (cb *ContextBuilder) SetPackageFilter(filter string) {
	cb.packageFilter = filter
}

// BuildContext scans the repository and creates comprehensive context.
func (cb *ContextBuilder) BuildContext(ctx context.Context) (*models.RepoContext, error) {
	if err := os.MkdirAll(cb.contextDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create context directory: %w", err)
	}

	packages, err := cb.scanAllFiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to scan files: %w", err)
	}

	var allFunctions []models.FunctionDetail
	var packageDetails []models.PackageDetail
	totalTests := 0

	for pkgPath, pkgInfo := range packages {
		pkgDetail := models.PackageDetail{
			Name:      pkgInfo.Name,
			Path:      pkgPath,
			Functions: []models.FunctionDetail{},
			TestFiles: pkgInfo.TestFiles,
		}

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
				Params:          fn.Params,
				Returns:         fn.Returns,
				HasTests:        false,
				ExistingTests:   []string{},
			}

			existingTests := findTestsForFunction(fn.Name, pkgInfo.TestFunctions)
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

	repoContext := &models.RepoContext{
		Repository: cb.repoPath,
		CreatedAt:  time.Now(),
		Summary: models.ContextSummary{
			TotalPackages:  len(packageDetails),
			TotalFunctions: len(allFunctions),
			TotalTests:     totalTests,
		},
		Packages:     packageDetails,
		AllFunctions: allFunctions,
	}

	if err := cb.saveContext(repoContext); err != nil {
		return nil, fmt.Errorf("failed to save context: %w", err)
	}

	return repoContext, nil
}

// LoadContext loads previously saved context from disk.
func (cb *ContextBuilder) LoadContext(_ context.Context) (*models.RepoContext, error) {
	data, err := os.ReadFile(filepath.Join(cb.contextDir, "context.json"))
	if err != nil {
		return nil, err
	}

	var repoContext models.RepoContext
	if err := json.Unmarshal(data, &repoContext); err != nil {
		return nil, err
	}
	return &repoContext, nil
}

func (cb *ContextBuilder) scanAllFiles(ctx context.Context) (map[string]*packageInfo, error) {
	packages := make(map[string]*packageInfo)

	err := filepath.WalkDir(cb.repoPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if strings.Contains(path, contextDirName) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if shouldSkipPath(path) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		pkgPath := getPackagePath(cb.repoPath, path)

		if cb.packageFilter != "" && !strings.Contains(pkgPath, cb.packageFilter) {
			return nil
		}

		pkgName := getPackageName(path)

		if _, exists := packages[pkgPath]; !exists {
			packages[pkgPath] = &packageInfo{
				Name:          pkgName,
				Functions:     []models.FunctionInfo{},
				TestFunctions: []string{},
				TestFiles:     []string{},
			}
		}

		if strings.HasSuffix(path, "_test.go") {
			testFuncs, err := scanner.ParseFile(path)
			if err == nil {
				for _, fn := range testFuncs {
					packages[pkgPath].TestFunctions = append(packages[pkgPath].TestFunctions, fn.Name)
				}
				packages[pkgPath].TestFiles = append(packages[pkgPath].TestFiles, path)
			}
		} else {
			functions, err := scanner.ParseFile(path)
			if err == nil {
				packages[pkgPath].Functions = append(packages[pkgPath].Functions, functions...)
			}
		}

		return nil
	})

	return packages, err
}

func (cb *ContextBuilder) saveContext(repoContext *models.RepoContext) error {
	data, err := json.MarshalIndent(repoContext, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(cb.contextDir, "context.json"), data, 0644)
}

type packageInfo struct {
	Name          string
	Functions     []models.FunctionInfo
	TestFunctions []string
	TestFiles     []string
}

// findTestsForFunction finds test functions that test a given function.
func findTestsForFunction(funcName string, testFunctions []string) []string {
	patterns := []string{
		"Test" + funcName,
		"Test_" + funcName,
		"TestNew" + funcName,
	}

	var matches []string
	for _, testFunc := range testFunctions {
		for _, pattern := range patterns {
			if strings.Contains(testFunc, pattern) {
				matches = append(matches, testFunc)
				break
			}
		}
	}
	return matches
}

func shouldSkipPath(path string) bool {
	skipDirs := []string{"vendor/", ".git/", "node_modules/", ".idea/", ".vscode/"}
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

// getPackageName reads the actual Go package declaration from the file's directory.
// Falls back to the directory name if parsing fails.
func getPackageName(filePath string) string {
	dir := filepath.Dir(filePath)

	entries, err := os.ReadDir(dir)
	if err == nil {
		fset := token.NewFileSet()
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
				continue
			}
			f, err := parser.ParseFile(fset, filepath.Join(dir, entry.Name()), nil, parser.PackageClauseOnly)
			if err == nil && f.Name != nil {
				return f.Name.Name
			}
		}
	}

	return filepath.Base(dir)
}

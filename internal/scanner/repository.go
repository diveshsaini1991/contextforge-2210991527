package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/divesh/contextforge/internal/models"
)

// ScanRepository scans a Go repository and extracts function-level metadata
func ScanRepository(repoPath string) (*models.RepositoryContext, error) {
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return nil, err
	}

	packages := make(map[string]*models.PackageInfo)

	err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-Go files and test files
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Skip vendor and hidden directories
		if shouldSkipPath(path) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Parse the file
		functions, err := ParseFile(path)
		if err != nil {
			// Log error but continue scanning
			return nil
		}

		if len(functions) == 0 {
			return nil
		}

		// Determine package name and path
		pkgPath := getPackagePath(absPath, path)
		pkgName := getPackageName(path)

		// Add to packages map
		if pkg, exists := packages[pkgPath]; exists {
			pkg.Functions = append(pkg.Functions, functions...)
		} else {
			packages[pkgPath] = &models.PackageInfo{
				Name:      pkgName,
				Path:      pkgPath,
				Functions: functions,
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Convert map to slice
	var pkgList []models.PackageInfo
	totalFunctions := 0
	for _, pkg := range packages {
		pkgList = append(pkgList, *pkg)
		totalFunctions += len(pkg.Functions)
	}

	context := &models.RepositoryContext{
		Repository:     absPath,
		ScannedAt:      time.Now(),
		Packages:       pkgList,
		TotalFunctions: totalFunctions,
		TotalPackages:  len(pkgList),
	}

	return context, nil
}

// shouldSkipPath determines if a path should be skipped during scanning
func shouldSkipPath(path string) bool {
	skipDirs := []string{
		"vendor/",
		".git/",
		"node_modules/",
		".idea/",
		".vscode/",
	}

	for _, skip := range skipDirs {
		if strings.Contains(path, skip) {
			return true
		}
	}

	return false
}

// getPackagePath extracts the relative package path from a file path
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

// getPackageName extracts the package name from a directory path
func getPackageName(filePath string) string {
	dir := filepath.Dir(filePath)
	return filepath.Base(dir)
}

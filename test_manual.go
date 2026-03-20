// Simple manual test to verify ContextForge functionality
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/divesh/contextforge/internal/coverage"
	"github.com/divesh/contextforge/internal/generator"
	"github.com/divesh/contextforge/internal/models"
	"github.com/divesh/contextforge/internal/scanner"
)

func main() {
	repoPath := "../test-repo"

	fmt.Println("=== Phase 1: Repository Scanning ===")
	ctx, err := scanner.ScanRepository(repoPath)
	if err != nil {
		log.Fatalf("Failed to scan repository: %v", err)
	}

	fmt.Printf("Repository: %s\n", ctx.Repository)
	fmt.Printf("Total Packages: %d\n", ctx.TotalPackages)
	fmt.Printf("Total Functions: %d\n\n", ctx.TotalFunctions)

	for _, pkg := range ctx.Packages {
		fmt.Printf("Package: %s (%s)\n", pkg.Name, pkg.Path)
		for _, fn := range pkg.Functions {
			exported := ""
			if fn.Exported {
				exported = "[EXPORTED]"
			}
			fmt.Printf("  - %s %s (lines: %d, complexity: %d) %s\n",
				fn.Name, fn.Signature, fn.LineCount, fn.ComplexityScore, exported)
		}
		fmt.Println()
	}

	fmt.Println("\n=== Phase 2: Coverage Analysis ===")
	analyzer := coverage.NewAnalyzer(repoPath, ctx)
	report, err := analyzer.Analyze()
	if err != nil {
		log.Fatalf("Failed to analyze coverage: %v", err)
	}

	fmt.Printf("Overall Coverage: %.2f%%\n", report.OverallCoverage)
	fmt.Printf("Fully Covered: %d\n", report.TotalFullyCovered)
	fmt.Printf("Partially Covered: %d\n", report.TotalPartiallyCovered)
	fmt.Printf("Uncovered: %d\n\n", report.TotalUncovered)

	fmt.Println("\n=== Phase 3: Smart Prioritization ===")
	var allUncovered []models.FunctionCoverage
	for _, pkg := range report.CoverageByPackage {
		allUncovered = append(allUncovered, pkg.UncoveredFunctions...)
	}

	prioritizer := generator.NewPrioritizer()
	topUncovered := prioritizer.GetTopPriority(allUncovered, 5)

	fmt.Println("Top 5 Uncovered Functions:")
	for i, fn := range topUncovered {
		fmt.Printf("%d. %s (Score: %d, Complexity: %d, Lines: %d)\n",
			i+1, fn.Name, fn.PriorityScore, fn.ComplexityScore, fn.LineCount)
	}

	// Output JSON for testing MCP integration
	fmt.Println("\n=== JSON Output (for MCP testing) ===")
	jsonData, _ := json.MarshalIndent(topUncovered, "", "  ")
	fmt.Println(string(jsonData))
}

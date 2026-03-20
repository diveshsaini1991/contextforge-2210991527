package generator

import (
	"sort"
	"strings"

	"github.com/divesh/contextforge/internal/models"
)

// Prioritizer handles smart prioritization of uncovered functions
type Prioritizer struct{}

// NewPrioritizer creates a new prioritizer
func NewPrioritizer() *Prioritizer {
	return &Prioritizer{}
}

// PrioritizeFunctions calculates priority scores and sorts functions
func (p *Prioritizer) PrioritizeFunctions(functions []models.FunctionCoverage) []models.FunctionCoverage {
	// Calculate priority scores
	for i := range functions {
		functions[i].PriorityScore = p.calculatePriorityScore(functions[i])
	}

	// Sort by priority score (descending)
	sort.Slice(functions, func(i, j int) bool {
		return functions[i].PriorityScore > functions[j].PriorityScore
	})

	return functions
}

// calculatePriorityScore calculates priority score based on multiple factors
func (p *Prioritizer) calculatePriorityScore(fn models.FunctionCoverage) int {
	score := 0

	// Exported (public API): +50 points
	if fn.Exported {
		score += 50
	}

	// Complexity score: complexity × 10 points
	score += fn.ComplexityScore * 10

	// Line count: min(lines, 50) points
	lineScore := fn.LineCount
	if lineScore > 50 {
		lineScore = 50
	}
	score += lineScore

	// Package importance
	score += p.getPackageImportance(fn.File)

	return score
}

// getPackageImportance assigns importance score based on package location
func (p *Prioritizer) getPackageImportance(filePath string) int {
	switch {
	case strings.Contains(filePath, "/cmd/") || strings.Contains(filePath, "/main"):
		return 20 // Main packages
	case strings.Contains(filePath, "/pkg/"):
		return 20 // Public packages
	case strings.Contains(filePath, "/internal/"):
		return 10 // Internal packages
	case strings.Contains(filePath, "/api/"):
		return 20 // API packages
	default:
		return 5 // Other packages
	}
}

// GetTopPriority returns the top N highest priority functions
func (p *Prioritizer) GetTopPriority(functions []models.FunctionCoverage, limit int) []models.FunctionCoverage {
	prioritized := p.PrioritizeFunctions(functions)

	if limit > 0 && limit < len(prioritized) {
		return prioritized[:limit]
	}

	return prioritized
}

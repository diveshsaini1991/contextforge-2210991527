package models

import "time"

// CoverageStatus represents the coverage state of a function
type CoverageStatus string

const (
	CoverageUncovered CoverageStatus = "uncovered"
	CoveragePartial   CoverageStatus = "partial"
	CoverageFull      CoverageStatus = "full"
)

// FunctionCoverage represents coverage information for a single function
type FunctionCoverage struct {
	FunctionInfo
	CurrentCoverage float64        `json:"current_coverage"`
	Status          CoverageStatus `json:"status"`
	PriorityScore   int            `json:"priority_score"`
}

// PackageCoverage represents coverage for a package
type PackageCoverage struct {
	Package            string             `json:"package"`
	CoveragePercent    float64            `json:"coverage_percent"`
	UncoveredFunctions []FunctionCoverage `json:"uncovered_functions"`
	PartiallyCovered   []FunctionCoverage `json:"partially_covered"`
	FullyCovered       []FunctionCoverage `json:"fully_covered"`
}

// CoverageReport represents the complete coverage analysis
type CoverageReport struct {
	AnalyzedAt            time.Time         `json:"analyzed_at"`
	OverallCoverage       float64           `json:"overall_coverage"`
	CoverageByPackage     []PackageCoverage `json:"coverage_by_package"`
	TotalUncovered        int               `json:"total_uncovered"`
	TotalPartiallyCovered int               `json:"total_partially_covered"`
	TotalFullyCovered     int               `json:"total_fully_covered"`
}

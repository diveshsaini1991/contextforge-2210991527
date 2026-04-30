package models

import "time"

// RepoContext represents human-readable repository documentation
type RepoContext struct {
	Repository    string          `json:"repository"`
	CreatedAt     time.Time       `json:"created_at"`
	Summary       ContextSummary  `json:"summary"`
	Packages      []PackageDetail `json:"packages"`
	AllFunctions  []FunctionDetail `json:"all_functions"`
}

// ContextSummary provides high-level repo statistics
type ContextSummary struct {
	TotalPackages  int `json:"total_packages"`
	TotalFunctions int `json:"total_functions"`
	TotalTests     int `json:"total_tests"`
}

// PackageDetail contains detailed package information
type PackageDetail struct {
	Name      string           `json:"name"`
	Path      string           `json:"path"`
	Functions []FunctionDetail `json:"functions"`
	TestFiles []string         `json:"test_files"`
}

// FunctionDetail represents a function with full context
type FunctionDetail struct {
	ID              string       `json:"id"` // package.FunctionName
	Name            string       `json:"name"`
	Package         string       `json:"package"`
	Signature       string       `json:"signature"`
	Exported        bool         `json:"exported"`
	FilePath        string       `json:"file_path"`
	StartLine       int          `json:"start_line"`
	EndLine         int          `json:"end_line"`
	LineCount       int          `json:"line_count"`
	ComplexityScore int          `json:"complexity_score"`
	ReceiverType    string       `json:"receiver_type,omitempty"`
	Params          []ParamInfo  `json:"params,omitempty"`
	Returns         []ReturnInfo `json:"returns,omitempty"`
	HasTests        bool         `json:"has_tests"`
	ExistingTests   []string     `json:"existing_tests"`
}

// TestScenario represents a test case that should exist
type TestScenario struct {
	ID           string `json:"id"` // Unique scenario ID
	FunctionID   string `json:"function_id"`
	FunctionName string `json:"function_name"`
	Package      string `json:"package"`
	ScenarioType string `json:"scenario_type"` // "happy_path", "edge_case", "error_case", "boundary"
	Description  string `json:"description"`
	TestName     string `json:"test_name"` // Suggested test function name
	Exists       bool   `json:"exists"`
	FoundIn      string `json:"found_in,omitempty"` // Test file where it was found
}

// ScenarioAnalysis contains all test scenarios for a repository
type ScenarioAnalysis struct {
	Repository     string         `json:"repository"`
	AnalyzedAt     time.Time      `json:"analyzed_at"`
	TotalScenarios int            `json:"total_scenarios"`
	Scenarios      []TestScenario `json:"scenarios"`
}

// TestCoverageReport shows what tests exist vs what's missing
type TestCoverageReport struct {
	Repository         string          `json:"repository"`
	GeneratedAt        time.Time       `json:"generated_at"`
	TotalFunctions     int             `json:"total_functions"`
	TestedFunctions    int             `json:"tested_functions"`
	UntestedFunctions  int             `json:"untested_functions"`
	FunctionCoverage   float64         `json:"function_coverage_percent"`
	TotalScenarios     int             `json:"total_scenarios"`
	CoveredScenarios   int             `json:"covered_scenarios"`
	MissingScenarios   int             `json:"missing_scenarios"`
	ScenarioCoverage   float64         `json:"scenario_coverage_percent"`
	ByPackage          []PackageReport `json:"by_package"`
	UntestedFuncList   []string        `json:"untested_functions_list"`
	MissingTests       []TestScenario  `json:"missing_tests"`
	Summary            string          `json:"summary"`
}

// PackageReport shows coverage for a single package
type PackageReport struct {
	Package          string  `json:"package"`
	TotalScenarios   int     `json:"total_scenarios"`
	CoveredScenarios int     `json:"covered_scenarios"`
	CoveragePercent  float64 `json:"coverage_percent"`
}

package models

import "time"

// FunctionInfo represents metadata about a function in the codebase
type FunctionInfo struct {
	Name            string `json:"name"`
	Signature       string `json:"signature"`
	Exported        bool   `json:"exported"`
	File            string `json:"file"`
	StartLine       int    `json:"start_line"`
	EndLine         int    `json:"end_line"`
	LineCount       int    `json:"line_count"`
	ComplexityScore int    `json:"complexity_score"`
	ReceiverType    string `json:"receiver_type,omitempty"` // For methods
}

// PackageInfo represents a Go package with its functions
type PackageInfo struct {
	Name      string         `json:"name"`
	Path      string         `json:"path"`
	Functions []FunctionInfo `json:"functions"`
}

// RepositoryContext represents the complete repository analysis
type RepositoryContext struct {
	Repository     string        `json:"repository"`
	ScannedAt      time.Time     `json:"scanned_at"`
	Packages       []PackageInfo `json:"packages"`
	TotalFunctions int           `json:"total_functions"`
	TotalPackages  int           `json:"total_packages"`
}

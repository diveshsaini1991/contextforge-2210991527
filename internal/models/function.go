package models

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
	ReceiverType    string `json:"receiver_type,omitempty"`
}

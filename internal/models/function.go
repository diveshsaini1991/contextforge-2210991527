package models

// ParamInfo represents a function parameter.
type ParamInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// ReturnInfo represents a function return type.
type ReturnInfo struct {
	Type string `json:"type"`
}

// FunctionInfo represents metadata about a function in the codebase.
type FunctionInfo struct {
	Name            string       `json:"name"`
	Signature       string       `json:"signature"`
	Exported        bool         `json:"exported"`
	File            string       `json:"file"`
	StartLine       int          `json:"start_line"`
	EndLine         int          `json:"end_line"`
	LineCount       int          `json:"line_count"`
	ComplexityScore int          `json:"complexity_score"`
	ReceiverType    string       `json:"receiver_type,omitempty"`
	Params          []ParamInfo  `json:"params,omitempty"`
	Returns         []ReturnInfo `json:"returns,omitempty"`
}

package generator

import (
	"fmt"

	"github.com/divesh/contextforge/internal/models"
)

// TestGenerator generates context for AI-powered test generation
type TestGenerator struct {
	repoPath string
}

// NewTestGenerator creates a new test generator
func NewTestGenerator(repoPath string) *TestGenerator {
	return &TestGenerator{
		repoPath: repoPath,
	}
}

// GenerateTestContext extracts context for the AI to generate tests
// Returns test context that should be passed to the AI assistant
func (g *TestGenerator) GenerateTestContext(functions []models.FunctionCoverage) (*TestContext, error) {
	extractor := NewContextExtractor(g.repoPath)
	return extractor.ExtractTestContext(functions)
}

// TestGenerationResult represents the result of test context extraction
type TestGenerationResult struct {
	TestContext  *TestContext `json:"test_context"`
	Message      string       `json:"message"`
	Instructions string       `json:"instructions_for_ai"`
}

// PrepareTestGeneration prepares everything needed for AI-driven test generation
func (g *TestGenerator) PrepareTestGeneration(functions []models.FunctionCoverage) (*TestGenerationResult, error) {
	if len(functions) == 0 {
		return nil, fmt.Errorf("no functions provided for test generation")
	}

	// Extract comprehensive context
	context, err := g.GenerateTestContext(functions)
	if err != nil {
		return nil, fmt.Errorf("failed to extract context: %w", err)
	}

	result := &TestGenerationResult{
		TestContext: context,
		Message: fmt.Sprintf(
			"Ready to generate tests for %d function(s) in package '%s'. "+
				"Test file will be created at: %s",
			len(functions),
			context.PackageName,
			context.TestFilePath,
		),
		Instructions: "Use the provided context to generate comprehensive unit tests. " +
			"Write the complete test file to the specified test_file_path. " +
			"Follow the detailed instructions in the test_context.",
	}

	return result, nil
}

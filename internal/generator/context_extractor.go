package generator

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/divesh/contextforge/internal/models"
)

// TestContext represents all context needed for test generation
type TestContext struct {
	PackageName     string                      `json:"package_name"`
	PackagePath     string                      `json:"package_path"`
	TestFilePath    string                      `json:"test_file_path"`
	Imports         []string                    `json:"imports"`
	TypeDefinitions string                      `json:"type_definitions"`
	Functions       []FunctionContext           `json:"functions"`
	RelatedCode     string                      `json:"related_code"`
	Instructions    string                      `json:"instructions"`
}

// FunctionContext represents context for a single function
type FunctionContext struct {
	Name       string `json:"name"`
	Signature  string `json:"signature"`
	SourceCode string `json:"source_code"`
	Complexity int    `json:"complexity"`
	LineCount  int    `json:"line_count"`
}

// ContextExtractor extracts comprehensive context for test generation
type ContextExtractor struct {
	repoPath string
}

// NewContextExtractor creates a new context extractor
func NewContextExtractor(repoPath string) *ContextExtractor {
	return &ContextExtractor{
		repoPath: repoPath,
	}
}

// ExtractTestContext extracts all context needed for AI to generate tests
func (e *ContextExtractor) ExtractTestContext(functions []models.FunctionCoverage) (*TestContext, error) {
	if len(functions) == 0 {
		return nil, fmt.Errorf("no functions provided")
	}

	// Group by package (assume all functions are from same package for now)
	pkgDir := filepath.Dir(functions[0].File)
	pkgName := filepath.Base(pkgDir)

	context := &TestContext{
		PackageName:  pkgName,
		PackagePath:  pkgDir,
		TestFilePath: filepath.Join(pkgDir, "contextforge_generated_test.go"),
		Functions:    []FunctionContext{},
	}

	// Extract source file context
	sourceFile := functions[0].File
	content, err := os.ReadFile(sourceFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read source file: %w", err)
	}

	// Parse the file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, sourceFile, content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source file: %w", err)
	}

	// Extract imports
	context.Imports = e.extractImports(node)

	// Extract type definitions
	context.TypeDefinitions = e.extractTypeDefinitions(node, fset, content)

	// Extract function source code
	lines := strings.Split(string(content), "\n")
	for _, fn := range functions {
		if fn.StartLine > 0 && fn.EndLine <= len(lines) {
			functionCode := strings.Join(lines[fn.StartLine-1:fn.EndLine], "\n")
			context.Functions = append(context.Functions, FunctionContext{
				Name:       fn.Name,
				Signature:  fn.Signature,
				SourceCode: functionCode,
				Complexity: fn.ComplexityScore,
				LineCount:  fn.LineCount,
			})
		}
	}

	// Extract related code (constants, variables)
	context.RelatedCode = e.extractRelatedCode(lines)

	// Add instructions for AI
	context.Instructions = e.generateInstructions(context)

	return context, nil
}

// extractImports extracts import statements
func (e *ContextExtractor) extractImports(node *ast.File) []string {
	var imports []string
	for _, imp := range node.Imports {
		imports = append(imports, imp.Path.Value)
	}
	return imports
}

// extractTypeDefinitions extracts type definitions from the AST
func (e *ContextExtractor) extractTypeDefinitions(node *ast.File, fset *token.FileSet, content []byte) string {
	var typeDefsBuilder strings.Builder

	ast.Inspect(node, func(n ast.Node) bool {
		switch t := n.(type) {
		case *ast.GenDecl:
			if t.Tok == token.TYPE {
				start := fset.Position(t.Pos()).Offset
				end := fset.Position(t.End()).Offset
				if start >= 0 && end <= len(content) {
					typeDefsBuilder.WriteString(string(content[start:end]))
					typeDefsBuilder.WriteString("\n\n")
				}
			}
		}
		return true
	})

	return typeDefsBuilder.String()
}

// extractRelatedCode extracts constants and variables
func (e *ContextExtractor) extractRelatedCode(lines []string) string {
	var relatedBuilder strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "const ") || strings.HasPrefix(trimmed, "var ") {
			relatedBuilder.WriteString(line)
			relatedBuilder.WriteString("\n")
		}
	}

	return relatedBuilder.String()
}

// generateInstructions creates detailed instructions for the AI
func (e *ContextExtractor) generateInstructions(ctx *TestContext) string {
	var funcList strings.Builder
	for _, fn := range ctx.Functions {
		funcList.WriteString(fmt.Sprintf("  - %s (complexity: %d, lines: %d)\n", fn.Signature, fn.Complexity, fn.LineCount))
	}

	instructions := fmt.Sprintf(`INSTRUCTIONS FOR TEST GENERATION:

Package: %s
Test File: %s

Functions to test:
%s

Context provided:
- Imports: All package imports
- Type Definitions: Structs, interfaces, and custom types
- Function Source Code: Complete implementation of each function
- Related Code: Package-level constants and variables

Requirements for generated tests:

1. STRUCTURE:
   - Start with "package %s"
   - Add necessary imports (testing, httptest, etc.)
   - Use table-driven tests for functions with multiple cases
   - One test function per target function (e.g., TestFunctionName)

2. TEST CASES:
   - Include happy path scenarios
   - Include edge cases (empty inputs, nil values, boundary conditions)
   - Include error cases (invalid inputs, expected errors)
   - For functions returning errors, test both success and failure paths

3. ASSERTIONS:
   - Use t.Errorf() for comparison failures
   - Use t.Fatalf() for fatal errors that prevent further testing
   - Check both returned values and error conditions
   - Verify state changes if applicable

4. MOCKING/TEST DOUBLES:
   - For gin.Context: Use gin.CreateTestContext() and httptest.ResponseRecorder
   - For HTTP handlers: Use httptest package
   - For database dependencies: Use in-memory implementations or simple mocks
   - For external services: Create simple stub implementations

5. DO NOT:
   - Do not use t.Skip() - write actual test implementations
   - Do not leave TODO comments - complete all tests
   - Do not import unnecessary packages
   - Do not use complex mocking libraries unless absolutely necessary

6. FORMAT:
   - Use clear, descriptive test case names
   - Add brief comments for complex test logic
   - Follow Go testing conventions
   - Keep tests simple and readable

OUTPUT:
Generate ONLY the complete Go test file content. No explanations, no markdown code fences.
Just the raw Go code ready to be written to %s`, ctx.PackageName, ctx.TestFilePath, funcList.String(), ctx.PackageName, ctx.TestFilePath)

	return instructions
}

// ToJSON converts the context to JSON for easy consumption
func (ctx *TestContext) ToJSON() (string, error) {
	data, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

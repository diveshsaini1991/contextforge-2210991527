package scanner

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"unicode"

	"github.com/divesh/contextforge/internal/models"
)

// ParseFile parses a Go source file and extracts function information
func ParseFile(filePath string) ([]models.FunctionInfo, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var functions []models.FunctionInfo

	ast.Inspect(node, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		funcInfo := extractFunctionInfo(fset, funcDecl, filePath)
		functions = append(functions, funcInfo)
		return true
	})

	return functions, nil
}

// extractFunctionInfo extracts metadata from a function declaration
func extractFunctionInfo(fset *token.FileSet, funcDecl *ast.FuncDecl, filePath string) models.FunctionInfo {
	funcName := funcDecl.Name.Name
	exported := isExported(funcName)

	startPos := fset.Position(funcDecl.Pos())
	endPos := fset.Position(funcDecl.End())

	signature := buildSignature(funcDecl)
	complexity := calculateComplexity(funcDecl)

	funcInfo := models.FunctionInfo{
		Name:            funcName,
		Signature:       signature,
		Exported:        exported,
		File:            filePath,
		StartLine:       startPos.Line,
		EndLine:         endPos.Line,
		LineCount:       endPos.Line - startPos.Line + 1,
		ComplexityScore: complexity,
	}

	// Check if this is a method (has a receiver)
	if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
		funcInfo.ReceiverType = getReceiverType(funcDecl.Recv.List[0].Type)
	}

	return funcInfo
}

// isExported checks if a function name is exported (public)
func isExported(name string) bool {
	if name == "" {
		return false
	}
	return unicode.IsUpper(rune(name[0]))
}

// buildSignature constructs the function signature string
func buildSignature(funcDecl *ast.FuncDecl) string {
	var sb strings.Builder

	sb.WriteString("func ")

	// Add receiver if present (for methods)
	if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
		sb.WriteString("(")
		sb.WriteString(getReceiverType(funcDecl.Recv.List[0].Type))
		sb.WriteString(") ")
	}

	sb.WriteString(funcDecl.Name.Name)
	sb.WriteString("(")

	// Add parameters
	if funcDecl.Type.Params != nil {
		for i, field := range funcDecl.Type.Params.List {
			if i > 0 {
				sb.WriteString(", ")
			}
			for j, name := range field.Names {
				if j > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(name.Name)
			}
			if len(field.Names) > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(exprToString(field.Type))
		}
	}

	sb.WriteString(")")

	// Add return type(s)
	if funcDecl.Type.Results != nil && len(funcDecl.Type.Results.List) > 0 {
		sb.WriteString(" ")
		if len(funcDecl.Type.Results.List) > 1 {
			sb.WriteString("(")
		}
		for i, field := range funcDecl.Type.Results.List {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(exprToString(field.Type))
		}
		if len(funcDecl.Type.Results.List) > 1 {
			sb.WriteString(")")
		}
	}

	return sb.String()
}

// getReceiverType extracts the receiver type from a method
func getReceiverType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.Ident:
		return t.Name
	default:
		return exprToString(t)
	}
}

// exprToString converts an expression to a string representation
func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + exprToString(t.Elt)
		}
		return "[" + exprToString(t.Len) + "]" + exprToString(t.Elt)
	case *ast.MapType:
		return "map[" + exprToString(t.Key) + "]" + exprToString(t.Value)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.Ellipsis:
		return "..." + exprToString(t.Elt)
	default:
		return "unknown"
	}
}

// calculateComplexity calculates cyclomatic complexity
func calculateComplexity(funcDecl *ast.FuncDecl) int {
	complexity := 1 // Base complexity

	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.IfStmt:
			complexity++
		case *ast.ForStmt:
			complexity++
		case *ast.RangeStmt:
			complexity++
		case *ast.SwitchStmt:
			complexity++
		case *ast.TypeSwitchStmt:
			complexity++
		case *ast.SelectStmt:
			complexity++
		case *ast.CaseClause:
			// Don't count default case
			if c, ok := n.(*ast.CaseClause); ok && c.List != nil {
				complexity++
			}
		case *ast.BinaryExpr:
			// Count && and || in conditions
			if b, ok := n.(*ast.BinaryExpr); ok {
				if b.Op == token.LAND || b.Op == token.LOR {
					complexity++
				}
			}
		}
		return true
	})

	return complexity
}

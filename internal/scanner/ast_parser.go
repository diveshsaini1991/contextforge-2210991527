package scanner

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"unicode"

	"github.com/diveshsaini1991/contextforge-2210991527/internal/models"
)

// ParseFile parses a Go source file and extracts function information.
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
		functions = append(functions, extractFunctionInfo(fset, funcDecl, filePath))
		return true
	})

	return functions, nil
}

func extractFunctionInfo(fset *token.FileSet, funcDecl *ast.FuncDecl, filePath string) models.FunctionInfo {
	startPos := fset.Position(funcDecl.Pos())
	endPos := fset.Position(funcDecl.End())

	funcInfo := models.FunctionInfo{
		Name:            funcDecl.Name.Name,
		Signature:       buildSignature(funcDecl),
		Exported:        isExported(funcDecl.Name.Name),
		File:            filePath,
		StartLine:       startPos.Line,
		EndLine:         endPos.Line,
		LineCount:       endPos.Line - startPos.Line + 1,
		ComplexityScore: calculateComplexity(funcDecl),
		Params:          extractParams(funcDecl),
		Returns:         extractReturns(funcDecl),
	}

	if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
		funcInfo.ReceiverType = getReceiverType(funcDecl.Recv.List[0].Type)
	}

	return funcInfo
}

func extractParams(funcDecl *ast.FuncDecl) []models.ParamInfo {
	if funcDecl.Type.Params == nil {
		return nil
	}

	var params []models.ParamInfo
	paramIdx := 0
	for _, field := range funcDecl.Type.Params.List {
		typStr := exprToString(field.Type)
		if len(field.Names) == 0 {
			params = append(params, models.ParamInfo{
				Name: fmt.Sprintf("arg%d", paramIdx),
				Type: typStr,
			})
			paramIdx++
		} else {
			for _, name := range field.Names {
				params = append(params, models.ParamInfo{
					Name: name.Name,
					Type: typStr,
				})
				paramIdx++
			}
		}
	}
	return params
}

func extractReturns(funcDecl *ast.FuncDecl) []models.ReturnInfo {
	if funcDecl.Type.Results == nil {
		return nil
	}

	var returns []models.ReturnInfo
	for _, field := range funcDecl.Type.Results.List {
		typStr := exprToString(field.Type)
		count := len(field.Names)
		if count == 0 {
			count = 1
		}
		for range count {
			returns = append(returns, models.ReturnInfo{Type: typStr})
		}
	}
	return returns
}

func isExported(name string) bool {
	if name == "" {
		return false
	}
	return unicode.IsUpper(rune(name[0]))
}

func buildSignature(funcDecl *ast.FuncDecl) string {
	var sb strings.Builder

	sb.WriteString("func ")

	if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
		sb.WriteString("(")
		sb.WriteString(getReceiverType(funcDecl.Recv.List[0].Type))
		sb.WriteString(") ")
	}

	sb.WriteString(funcDecl.Name.Name)
	sb.WriteString("(")

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
	case *ast.FuncType:
		return "func(...)"
	case *ast.ChanType:
		return "chan " + exprToString(t.Value)
	case *ast.BasicLit:
		return t.Value
	default:
		return "unknown"
	}
}

func calculateComplexity(funcDecl *ast.FuncDecl) int {
	complexity := 1

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
			if c, ok := n.(*ast.CaseClause); ok && c.List != nil {
				complexity++
			}
		case *ast.BinaryExpr:
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

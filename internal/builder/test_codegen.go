package builder

import (
	"fmt"
	"strings"

	"github.com/diveshsaini1991/contextforge-2210991527/internal/models"
)

type funcClass int

const (
	classTestable funcClass = iota
	classTODO
	classFallback
)

// GenerateTestCode produces a real test function body for the given function and scenario.
// Returns the test function code and a list of required imports beyond "testing".
// If the function can't be tested automatically, returns a TODO comment or t.Skip stub.
func GenerateTestCode(fn models.FunctionDetail, scenario models.TestScenario) (string, []string) {
	class := classifyFunction(fn)

	switch class {
	case classTODO:
		return generateTODO(fn, scenario), nil
	case classFallback:
		return generateSkipStub(scenario), nil
	default:
		return buildTest(fn, scenario)
	}
}

func classifyFunction(fn models.FunctionDetail) funcClass {
	if fn.Name == "main" || fn.Name == "init" {
		return classTODO
	}

	for _, p := range fn.Params {
		t := p.Type
		if strings.Contains(t, "gin.") ||
			strings.Contains(t, "http.ResponseWriter") ||
			strings.Contains(t, "http.Request") ||
			strings.Contains(t, "testing.T") ||
			strings.Contains(t, "testing.B") {
			return classTODO
		}
		if t == "unknown" || t == "func(...)" || strings.HasPrefix(t, "chan ") {
			return classFallback
		}
	}

	if len(fn.Params) == 0 && len(fn.Returns) == 0 {
		return classTODO
	}

	return classTestable
}

func generateTODO(fn models.FunctionDetail, scenario models.TestScenario) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n// TODO: %s\n", scenario.Description))

	hint := ""
	for _, p := range fn.Params {
		if strings.Contains(p.Type, "gin.Context") {
			hint = "// Requires gin test setup: use httptest.NewRecorder() and gin.CreateTestContext().\n"
			break
		}
		if strings.Contains(p.Type, "http.") {
			hint = "// Requires HTTP test setup: use httptest.NewRequest() and httptest.NewRecorder().\n"
			break
		}
	}
	if fn.Name == "main" || fn.Name == "init" {
		hint = "// Entry point function — consider integration testing instead.\n"
	}

	sb.WriteString(fmt.Sprintf("// func %s(t *testing.T) {\n", scenario.TestName))
	if hint != "" {
		sb.WriteString(hint)
	}
	sb.WriteString("// }\n")
	return sb.String()
}

func generateSkipStub(scenario models.TestScenario) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\nfunc %s(t *testing.T) {\n", scenario.TestName))
	sb.WriteString(fmt.Sprintf("\tt.Skip(%q)\n", "stub: "+scenario.Description))
	sb.WriteString("}\n")
	return sb.String()
}

func buildTest(fn models.FunctionDetail, scenario models.TestScenario) (string, []string) {
	var sb strings.Builder
	imports := map[string]bool{}

	hasError := fnReturnsError(fn)
	needsDeepEqual := returnsSliceOrMap(fn)

	if needsDeepEqual {
		imports["reflect"] = true
	}

	// Generate test cases
	cases := generateCases(fn, scenario, imports)

	// Build struct fields
	structFields := buildStructFields(fn, hasError)

	// Build function call
	callExpr, receiverSetup := buildFunctionCall(fn, imports)

	sb.WriteString(fmt.Sprintf("\nfunc %s(t *testing.T) {\n", scenario.TestName))
	sb.WriteString("\ttests := []struct {\n")
	sb.WriteString("\t\tname string\n")
	sb.WriteString(structFields)
	sb.WriteString("\t}{\n")
	sb.WriteString(cases)
	sb.WriteString("\t}\n")
	sb.WriteString("\tfor _, tt := range tests {\n")
	sb.WriteString("\t\tt.Run(tt.name, func(t *testing.T) {\n")

	if receiverSetup != "" {
		sb.WriteString("\t\t\t" + receiverSetup + "\n")
	}

	// Inject context.Background() params
	for _, p := range fn.Params {
		if p.Type == "context.Context" {
			imports["context"] = true
		}
	}

	// Build assertions
	nonErrorReturns := getNonErrorReturns(fn)

	if len(fn.Returns) == 0 {
		sb.WriteString(fmt.Sprintf("\t\t\t%s\n", callExpr))
	} else if hasError && len(nonErrorReturns) == 0 {
		// Only returns error
		sb.WriteString(fmt.Sprintf("\t\t\terr := %s\n", callExpr))
		sb.WriteString("\t\t\tif (err != nil) != tt.wantErr {\n")
		sb.WriteString(fmt.Sprintf("\t\t\t\tt.Errorf(\"%s() error = %%v, wantErr %%v\", err, tt.wantErr)\n", fn.Name))
		sb.WriteString("\t\t\t}\n")
	} else if hasError {
		sb.WriteString(fmt.Sprintf("\t\t\tgot, err := %s\n", callExpr))
		sb.WriteString("\t\t\tif (err != nil) != tt.wantErr {\n")
		sb.WriteString(fmt.Sprintf("\t\t\t\tt.Errorf(\"%s() error = %%v, wantErr %%v\", err, tt.wantErr)\n", fn.Name))
		sb.WriteString("\t\t\t\treturn\n")
		sb.WriteString("\t\t\t}\n")
		sb.WriteString("\t\t\tif !tt.wantErr {\n")
		if needsDeepEqual {
			sb.WriteString(fmt.Sprintf("\t\t\t\tif !reflect.DeepEqual(got, tt.want) {\n"))
			sb.WriteString(fmt.Sprintf("\t\t\t\t\tt.Errorf(\"%s() = %%v, want %%v\", got, tt.want)\n", fn.Name))
		} else {
			sb.WriteString(fmt.Sprintf("\t\t\t\tif got != tt.want {\n"))
			sb.WriteString(fmt.Sprintf("\t\t\t\t\tt.Errorf(\"%s() = %%v, want %%v\", got, tt.want)\n", fn.Name))
		}
		sb.WriteString("\t\t\t\t}\n")
		sb.WriteString("\t\t\t}\n")
	} else if len(nonErrorReturns) == 1 {
		sb.WriteString(fmt.Sprintf("\t\t\tgot := %s\n", callExpr))
		if needsDeepEqual {
			sb.WriteString("\t\t\tif !reflect.DeepEqual(got, tt.want) {\n")
		} else {
			sb.WriteString("\t\t\tif got != tt.want {\n")
		}
		sb.WriteString(fmt.Sprintf("\t\t\t\tt.Errorf(\"%s() = %%v, want %%v\", got, tt.want)\n", fn.Name))
		sb.WriteString("\t\t\t}\n")
	} else {
		// Multiple non-error returns — just call it
		sb.WriteString(fmt.Sprintf("\t\t\t_ = %s\n", callExpr))
	}

	sb.WriteString("\t\t})\n")
	sb.WriteString("\t}\n")
	sb.WriteString("}\n")

	var importList []string
	for imp := range imports {
		importList = append(importList, imp)
	}
	return sb.String(), importList
}

func buildStructFields(fn models.FunctionDetail, hasError bool) string {
	var sb strings.Builder

	for _, p := range fn.Params {
		if p.Type == "context.Context" {
			continue
		}
		sb.WriteString(fmt.Sprintf("\t\t%s %s\n", sanitizeFieldName(p.Name), p.Type))
	}

	nonErrorReturns := getNonErrorReturns(fn)
	if len(nonErrorReturns) == 1 {
		sb.WriteString(fmt.Sprintf("\t\twant %s\n", nonErrorReturns[0].Type))
	}

	if hasError {
		sb.WriteString("\t\twantErr bool\n")
	}

	return sb.String()
}

func buildFunctionCall(fn models.FunctionDetail, imports map[string]bool) (string, string) {
	var args []string
	for _, p := range fn.Params {
		if p.Type == "context.Context" {
			args = append(args, "context.Background()")
			imports["context"] = true
		} else {
			args = append(args, "tt."+sanitizeFieldName(p.Name))
		}
	}

	callExpr := fmt.Sprintf("%s(%s)", fn.Name, strings.Join(args, ", "))

	receiverSetup := ""
	if fn.ReceiverType != "" {
		varName := strings.ToLower(fn.ReceiverType[:1])
		varName = strings.TrimPrefix(varName, "*")
		if varName == "" {
			varName = "recv"
		}

		cleanType := strings.TrimPrefix(fn.ReceiverType, "*")
		if strings.HasPrefix(fn.ReceiverType, "*") {
			receiverSetup = fmt.Sprintf("%s := &%s{}", varName, cleanType)
		} else {
			receiverSetup = fmt.Sprintf("var %s %s", varName, cleanType)
		}
		callExpr = fmt.Sprintf("%s.%s(%s)", varName, fn.Name, strings.Join(args, ", "))
	}

	return callExpr, receiverSetup
}

func generateCases(fn models.FunctionDetail, scenario models.TestScenario, imports map[string]bool) string {
	var sb strings.Builder

	nonErrorReturns := getNonErrorReturns(fn)
	hasError := fnReturnsError(fn)

	switch scenario.ScenarioType {
	case "happy_path":
		sb.WriteString(generateHappyPathCases(fn, nonErrorReturns, hasError))
	case "error_case":
		sb.WriteString(generateErrorCases(fn, nonErrorReturns, hasError))
	case "boundary":
		sb.WriteString(generateBoundaryCases(fn, nonErrorReturns, hasError, imports))
	case "edge_case":
		sb.WriteString(generateEdgeCases(fn, nonErrorReturns, hasError))
	}

	return sb.String()
}

func generateHappyPathCases(fn models.FunctionDetail, nonErrorReturns []models.ReturnInfo, hasError bool) string {
	var sb strings.Builder

	caseNames := []string{"basic case", "another case", "third case"}
	valueSets := []int{0, 1, 2}

	for i, idx := range valueSets {
		if i >= 3 {
			break
		}
		sb.WriteString("\t\t{")
		sb.WriteString(fmt.Sprintf("%q, ", caseNames[i]))

		for j, p := range fn.Params {
			if p.Type == "context.Context" {
				continue
			}
			if j > 0 || countNonContextParams(fn) > 1 {
				if j > 0 {
					sb.WriteString(", ")
				}
			}
			sb.WriteString(sampleValue(p.Type, "happy_path", idx))
		}

		if len(nonErrorReturns) == 1 {
			sb.WriteString(", ")
			sb.WriteString(sampleValue(nonErrorReturns[0].Type, "happy_path", idx) + " /* TODO: set expected value */")
		}
		if hasError {
			sb.WriteString(", false")
		}
		sb.WriteString("},\n")
	}

	return sb.String()
}

func generateErrorCases(fn models.FunctionDetail, nonErrorReturns []models.ReturnInfo, hasError bool) string {
	var sb strings.Builder

	sb.WriteString("\t\t{")
	sb.WriteString("\"error case\", ")
	for j, p := range fn.Params {
		if p.Type == "context.Context" {
			continue
		}
		if j > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(sampleValue(p.Type, "error_case", 0))
	}
	if len(nonErrorReturns) == 1 {
		sb.WriteString(", " + zeroValue(nonErrorReturns[0].Type))
	}
	if hasError {
		sb.WriteString(", true")
	}
	sb.WriteString("},\n")

	sb.WriteString("\t\t{")
	sb.WriteString("\"valid case\", ")
	for j, p := range fn.Params {
		if p.Type == "context.Context" {
			continue
		}
		if j > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(sampleValue(p.Type, "happy_path", 0))
	}
	if len(nonErrorReturns) == 1 {
		sb.WriteString(", " + sampleValue(nonErrorReturns[0].Type, "happy_path", 0) + " /* TODO: set expected value */")
	}
	if hasError {
		sb.WriteString(", false")
	}
	sb.WriteString("},\n")

	return sb.String()
}

func generateBoundaryCases(fn models.FunctionDetail, nonErrorReturns []models.ReturnInfo, hasError bool, imports map[string]bool) string {
	var sb strings.Builder
	imports["math"] = true

	caseNames := []string{"zero values", "negative values", "large values"}
	indices := []int{0, 1, 2}

	for i, idx := range indices {
		sb.WriteString("\t\t{")
		sb.WriteString(fmt.Sprintf("%q, ", caseNames[i]))
		for j, p := range fn.Params {
			if p.Type == "context.Context" {
				continue
			}
			if j > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(sampleValue(p.Type, "boundary", idx))
		}
		if len(nonErrorReturns) == 1 {
			sb.WriteString(", " + zeroValue(nonErrorReturns[0].Type) + " /* TODO: set expected value */")
		}
		if hasError {
			if i > 0 {
				sb.WriteString(", false /* TODO: may error */")
			} else {
				sb.WriteString(", false")
			}
		}
		sb.WriteString("},\n")
	}

	return sb.String()
}

func generateEdgeCases(fn models.FunctionDetail, nonErrorReturns []models.ReturnInfo, hasError bool) string {
	var sb strings.Builder

	sb.WriteString("\t\t{")
	sb.WriteString("\"zero/empty inputs\", ")
	for j, p := range fn.Params {
		if p.Type == "context.Context" {
			continue
		}
		if j > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(zeroValue(p.Type))
	}
	if len(nonErrorReturns) == 1 {
		sb.WriteString(", " + zeroValue(nonErrorReturns[0].Type) + " /* TODO: set expected value */")
	}
	if hasError {
		sb.WriteString(", false /* TODO: may error */")
	}
	sb.WriteString("},\n")

	return sb.String()
}

func sampleValue(typStr string, scenarioType string, index int) string {
	switch scenarioType {
	case "happy_path":
		return happyValue(typStr, index)
	case "error_case":
		return errorValue(typStr, index)
	case "boundary":
		return boundaryValue(typStr, index)
	default:
		return zeroValue(typStr)
	}
}

func happyValue(typStr string, index int) string {
	intVals := []string{"1", "5", "10"}
	floatVals := []string{"1.5", "2.5", "3.14"}
	stringVals := []string{`"hello"`, `"world"`, `"test"`}
	boolVals := []string{"true", "false", "true"}

	switch {
	case typStr == "int" || typStr == "int64" || typStr == "int32" || typStr == "int16" || typStr == "int8":
		return intVals[index%len(intVals)]
	case typStr == "uint" || typStr == "uint64" || typStr == "uint32" || typStr == "uint16" || typStr == "uint8" || typStr == "byte":
		return intVals[index%len(intVals)]
	case typStr == "float64" || typStr == "float32":
		return floatVals[index%len(floatVals)]
	case typStr == "string":
		return stringVals[index%len(stringVals)]
	case typStr == "bool":
		return boolVals[index%len(boolVals)]
	case strings.HasPrefix(typStr, "[]"):
		elemType := strings.TrimPrefix(typStr, "[]")
		return typStr + "{" + happyValue(elemType, 0) + "}"
	case typStr == "error":
		return "nil"
	default:
		return zeroValue(typStr)
	}
}

func errorValue(typStr string, _ int) string {
	switch {
	case typStr == "int" || typStr == "int64" || typStr == "int32":
		return "0"
	case typStr == "uint" || typStr == "uint64" || typStr == "uint32":
		return "0"
	case typStr == "float64" || typStr == "float32":
		return "0"
	case typStr == "string":
		return `""`
	case typStr == "bool":
		return "false"
	case strings.HasPrefix(typStr, "[]") || strings.HasPrefix(typStr, "map[") || strings.HasPrefix(typStr, "*"):
		return "nil"
	default:
		return zeroValue(typStr)
	}
}

func boundaryValue(typStr string, index int) string {
	switch {
	case typStr == "int" || typStr == "int64":
		vals := []string{"0", "-1", "math.MaxInt32"}
		return vals[index%len(vals)]
	case typStr == "int32":
		vals := []string{"0", "-1", "math.MaxInt32"}
		return vals[index%len(vals)]
	case typStr == "uint" || typStr == "uint64" || typStr == "uint32":
		vals := []string{"0", "1", "math.MaxUint32"}
		return vals[index%len(vals)]
	case typStr == "float64" || typStr == "float32":
		vals := []string{"0.0", "-0.1", "math.MaxFloat64"}
		return vals[index%len(vals)]
	case typStr == "string":
		vals := []string{`""`, `"a"`, `"` + strings.Repeat("x", 100) + `"`}
		return vals[index%len(vals)]
	case typStr == "bool":
		return "false"
	case strings.HasPrefix(typStr, "[]"):
		vals := []string{"nil", typStr + "{}", typStr + "{" + happyValue(strings.TrimPrefix(typStr, "[]"), 0) + "}"}
		return vals[index%len(vals)]
	default:
		return zeroValue(typStr)
	}
}

func zeroValue(typStr string) string {
	switch {
	case typStr == "int" || typStr == "int64" || typStr == "int32" || typStr == "int16" || typStr == "int8":
		return "0"
	case typStr == "uint" || typStr == "uint64" || typStr == "uint32" || typStr == "uint16" || typStr == "uint8" || typStr == "byte":
		return "0"
	case typStr == "float64" || typStr == "float32":
		return "0.0"
	case typStr == "string":
		return `""`
	case typStr == "bool":
		return "false"
	case typStr == "error":
		return "nil"
	case strings.HasPrefix(typStr, "[]") || strings.HasPrefix(typStr, "map[") ||
		strings.HasPrefix(typStr, "*") || typStr == "interface{}":
		return "nil"
	default:
		return typStr + "{}"
	}
}

func fnReturnsError(fn models.FunctionDetail) bool {
	for _, r := range fn.Returns {
		if r.Type == "error" {
			return true
		}
	}
	return false
}

func returnsSliceOrMap(fn models.FunctionDetail) bool {
	for _, r := range fn.Returns {
		if r.Type == "error" {
			continue
		}
		if strings.HasPrefix(r.Type, "[]") || strings.HasPrefix(r.Type, "map[") {
			return true
		}
	}
	return false
}

func getNonErrorReturns(fn models.FunctionDetail) []models.ReturnInfo {
	var result []models.ReturnInfo
	for _, r := range fn.Returns {
		if r.Type != "error" {
			result = append(result, r)
		}
	}
	return result
}

func countNonContextParams(fn models.FunctionDetail) int {
	count := 0
	for _, p := range fn.Params {
		if p.Type != "context.Context" {
			count++
		}
	}
	return count
}

func sanitizeFieldName(name string) string {
	if name == "type" || name == "func" || name == "map" || name == "range" ||
		name == "return" || name == "select" || name == "case" || name == "default" {
		return name + "Val"
	}
	return name
}

# ContextForge MCP - Usage Guide

## Overview

ContextForge MCP creates repository context and generates test stubs efficiently. All data is cached in `.contextforge/` directory to avoid redundant scanning.

## Workflow

### Step 1: Build Repository Context
**Tool:** `build_repo_context`

Scans the entire repository and creates comprehensive documentation:
- All functions in the codebase
- Existing test files and test functions
- Which functions already have tests
- Package structure

**Output:** `.contextforge/context.json`

```json
{
  "repo_path": "/path/to/your/repo"
}
```

**What you get:**
- Human-readable documentation of your codebase
- Function signatures, complexity scores, line counts
- Existing test mapping (which functions are already tested)

---

### Step 2: Analyze Test Scenarios
**Tool:** `analyze_test_scenarios`

Generates a list of ALL test scenarios that should exist based on your code:
- Happy path tests (always)
- Error case tests (for functions returning errors)
- Edge case tests (for complex exported functions)
- Boundary tests (for functions with numeric/slice parameters)

**Output:** `.contextforge/scenarios.json`

```json
{
  "repo_path": "/path/to/your/repo"
}
```

**Uses existing context if available** - no need to rescan!

**What you get:**
- Complete list of test scenarios
- Suggested test function names
- Test descriptions
- Scenario types (happy_path, error_case, edge_case, boundary)

---

### Step 3: Check Test Coverage
**Tool:** `check_test_coverage`

Compares scenarios against existing tests and identifies gaps:
- Which scenarios are already covered
- Which scenarios are missing
- Coverage percentage by package

**Output:** `.contextforge/coverage_report.json`

```json
{
  "repo_path": "/path/to/your/repo"
}
```

**Uses existing context and scenarios** - very fast!

**What you get:**
- Test coverage report
- List of missing test scenarios
- Coverage percentage (overall and per package)
- Human-readable summary

---

### Step 4: Generate Test Stubs
**Tool:** `generate_test_stubs`

Creates test files with stub functions for all missing tests:
- Creates new test files if needed
- Appends to existing test files
- Each stub prints "test not yet built"
- Uses `t.Skip()` so tests pass but are marked as skipped

**Output:** Test files created/updated in your repository

```json
{
  "repo_path": "/path/to/your/repo"
}
```

**What you get:**
- Test files with stub functions
- Each stub contains:
  ```go
  // Test <description>
  func TestFunctionName(t *testing.T) {
      // TODO: Implement test logic
      t.Log("Test not yet built: <description>")
      t.Skip("Test stub - implementation pending")
  }
  ```

---

## Example Workflow

### Starting from scratch (no tests):

1. **Build context:**
   ```
   Tool: build_repo_context
   Result: .contextforge/context.json created
   Output: "Found 45 functions in 8 packages"
   ```

2. **Analyze scenarios:**
   ```
   Tool: analyze_test_scenarios
   Result: .contextforge/scenarios.json created
   Output: "Generated 127 test scenarios"
   ```

3. **Check coverage:**
   ```
   Tool: check_test_coverage
   Result: .contextforge/coverage_report.json created
   Output: "0 covered, 127 missing (0% coverage)"
   ```

4. **Generate stubs:**
   ```
   Tool: generate_test_stubs
   Result: Creates test files with 127 stub functions
   Output: "Generated 127 stubs in 8 test files"
   ```

---

### With existing tests:

1. **Build context:**
   ```
   Tool: build_repo_context
   Result: Finds existing test functions
   Output: "Found 45 functions, 23 existing tests"
   ```

2. **Analyze scenarios:**
   ```
   Tool: analyze_test_scenarios
   Result: Generates scenarios for all functions
   Output: "Generated 127 test scenarios"
   ```

3. **Check coverage:**
   ```
   Tool: check_test_coverage
   Result: Matches existing tests to scenarios
   Output: "35 covered, 92 missing (27.6% coverage)"
   ```

4. **Generate stubs:**
   ```
   Tool: generate_test_stubs
   Result: Creates only missing test stubs
   Output: "Generated 92 stubs in 6 test files"
   ```

---

## Efficiency Features

### Caching
- Context is saved to `.contextforge/context.json`
- Scenarios are saved to `.contextforge/scenarios.json`
- Coverage report is saved to `.contextforge/coverage_report.json`
- Subsequent tools reuse cached data

### Smart Matching
- Detects existing tests by name patterns
- Matches `TestFunctionName`, `Test_FunctionName`, etc.
- Identifies error tests, edge case tests automatically

### Incremental Updates
- Appends to existing test files
- Skips stubs that already exist
- Won't overwrite your work

---

## Tool: get_context

**Tool:** `get_context`

Retrieves the saved repository context without rebuilding.

```json
{
  "repo_path": "/path/to/your/repo"
}
```

**Use when:** You want to view the context without rescanning.

---

## File Structure

```
your-repo/
├── .contextforge/
│   ├── context.json         # Repository context
│   ├── scenarios.json       # Test scenarios
│   └── coverage_report.json # Coverage analysis
├── pkg/
│   ├── handler.go
│   └── handler_test.go      # Generated/updated by generate_test_stubs
└── internal/
    ├── service.go
    └── service_test.go      # Generated/updated by generate_test_stubs
```

---

## Benefits

1. **No redundant scanning** - Context is built once and reused
2. **Works with zero tests** - Generates complete test structure from scratch
3. **Works with existing tests** - Only adds missing tests
4. **Human-readable** - All JSON files are readable documentation
5. **Efficient** - Steps 2-4 use cached context, very fast
6. **No API keys needed** - Stub generation is template-based
7. **Non-destructive** - Appends to files, won't overwrite

---

## Next Steps (Future Implementation)

The stubs are created with `t.Skip()`. To implement the actual tests:

1. Remove `t.Skip()` from a stub
2. Add test implementation logic
3. Add assertions
4. Run tests

Later, we can add AI-powered test generation to fill in the stub logic automatically!

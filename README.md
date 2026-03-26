# ContextForge - Unit Test Coverage Intelligence System

ContextForge is an intelligent system that analyzes Go repositories to identify test coverage gaps and generate targeted unit tests using AI.

## Features

- **Function-Level Analysis**: Scans Go repositories and extracts detailed metadata about every function
- **Coverage Gap Detection**: Identifies uncovered, partially covered, and fully covered functions
- **Smart Prioritization**: Ranks uncovered functions based on:
  - Public API visibility (exported functions)
  - Cyclomatic complexity
  - Function size
  - Package importance
- **AI-Assisted Test Generation**: Extracts comprehensive context and leverages the AI assistant (Claude/Cursor) to generate real, working unit tests with:
  - Comprehensive test cases (happy path, edge cases, error cases)
  - Table-driven tests for complex functions
  - Proper mocking for dependencies (HTTP handlers, databases, etc.)
  - Meaningful assertions and error checking
  - **No API key required** - uses your existing Cursor/Claude subscription
- **MCP Integration**: Integrates seamlessly with Claude Desktop and Cursor via Model Context Protocol

## Architecture

ContextForge operates in three phases:

1. **Repository Context Generation**: Scans the codebase and builds a function-level map
2. **Coverage Analysis**: Runs `go test -cover` and identifies gaps
3. **Smart Test Generation**: Prioritizes and generates tests for uncovered functions

## Installation

```bash
cd contextforge
go mod download
go build -o contextforge ./cmd/mcp-server
```

## Usage

### As MCP Server

Add to your MCP settings:

**For Claude Desktop** (`~/Library/Application Support/Claude/claude_desktop_config.json`):
```json
{
  "mcpServers": {
    "contextforge": {
      "command": "/path/to/contextforge/contextforge"
    }
  }
}
```

**For Cursor** (`.cursorrules` or MCP settings):
```json
{
  "mcpServers": {
    "contextforge": {
      "command": "/path/to/contextforge/contextforge"
    }
  }
}
```

### Available MCP Tools

1. **build_repo_context**
   - Scans repository and creates comprehensive context
   - Saves to `.contextforge/context.json`
   - Input: `{"repo_path": "/path/to/repo"}`

2. **analyze_test_scenarios**
   - Generates all test scenarios that should exist
   - Saves to `.contextforge/scenarios.json`
   - Input: `{"repo_path": "/path/to/repo"}`

3. **check_test_coverage**
   - Compares scenarios vs existing tests
   - Saves to `.contextforge/coverage_report.json`
   - Input: `{"repo_path": "/path/to/repo"}`

4. **generate_test_stubs**
   - Creates test files with stub functions for missing tests
   - Each stub prints "test not yet built"
   - Input: `{"repo_path": "/path/to/repo"}`

5. **get_context**
   - Retrieves saved repository context
   - Input: `{"repo_path": "/path/to/repo"}`

See [MCP_USAGE.md](MCP_USAGE.md) for detailed workflow guide.

## How It Works

ContextForge uses an efficient 4-step workflow:

### Step 1: Build Repository Context
Scans your codebase once and creates comprehensive documentation:
- All functions with signatures and complexity
- Existing tests mapped to functions
- Saved to `.contextforge/context.json`

### Step 2: Analyze Test Scenarios
Generates test scenarios based on function characteristics:
- Happy path (always)
- Error cases (for functions returning errors)
- Edge cases (for complex functions)
- Boundary tests (for numeric/slice parameters)

### Step 3: Check Coverage
Compares scenarios against existing tests:
- Smart matching of test names
- Identifies coverage gaps
- Generates detailed report

### Step 4: Generate Test Stubs
Creates test files with stub functions:
- Each stub has `t.Skip("Test stub - implementation pending")`
- Appends to existing files or creates new ones
- Ready for implementation

**Advantage:** Efficient caching - context is built once and reused!

## Example Workflow

```
1. build_repo_context → Create .contextforge/context.json
2. analyze_test_scenarios → Generate .contextforge/scenarios.json
3. check_test_coverage → Generate .contextforge/coverage_report.json
4. generate_test_stubs → Create test files with stubs
```

## Prioritization Algorithm

Functions are scored based on:
- **Exported (public API)**: +50 points
- **Complexity**: complexity × 10 points
- **Line count**: min(lines, 50) points
- **Package importance**:
  - main/pkg/api: +20 points
  - internal: +10 points
  - other: +5 points

Higher scores = higher priority for test generation.

## Requirements

- Go 1.21 or higher
- Working Go module in the target repository
- Existing test infrastructure (go test must work)

## Project Structure

```
contextforge/
├── cmd/mcp-server/       # MCP server entry point
├── internal/
│   ├── scanner/          # Repository scanning & AST parsing
│   ├── coverage/         # Coverage analysis & profile parsing
│   ├── generator/        # Test generation & prioritization
│   └── models/           # Data structures
├── mcp/                  # MCP server implementation
└── README.md
```

## Limitations (MVP)

- Go language only
- Unit tests only (no integration tests)
- Single repository support
- Basic test templates (requires manual refinement)

## Future Enhancements

- Multi-language support (Python, TypeScript)
- Integration test generation
- LLM-powered intelligent test generation
- Coverage trend tracking
- CLI interface for CI/CD

## License

MIT


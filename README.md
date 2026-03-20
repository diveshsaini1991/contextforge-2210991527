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

**No API key needed!** The tool leverages your existing Claude/Cursor subscription.

### Available MCP Tools

1. **analyze_repository**
   - Scans a Go repository and extracts function metadata
   - Input: `{"repo_path": "/path/to/repo"}`

2. **analyze_coverage**
   - Runs coverage analysis and identifies gaps
   - Input: `{"repo_path": "/path/to/repo"}`

3. **get_test_recommendations**
   - Returns prioritized list of uncovered functions
   - Input: `{"repo_path": "/path/to/repo", "limit": 10}`

4. **generate_unit_tests**
   - Generates tests for specified functions
   - Input: `{"repo_path": "/path/to/repo", "function_ids": ["FuncName1", "FuncName2"]}`

5. **verify_coverage_improvement**
   - Re-runs coverage to verify improvements
   - Input: `{"repo_path": "/path/to/repo"}`

## How It Works

ContextForge uses a two-step AI-assisted approach:

### Step 1: Context Extraction
The MCP server extracts comprehensive context for each function:
- Function source code
- All imports
- Type definitions (structs, interfaces)
- Related code (constants, variables)
- Detailed generation instructions

### Step 2: AI-Powered Generation
The AI assistant (Claude/Cursor) receives this context and generates **real, working tests**:
- ✅ Comprehensive test cases with assertions
- ✅ Edge cases and error handling
- ✅ Table-driven tests for complex functions
- ✅ Proper mocking for dependencies (gin.Context, databases, etc.)
- ✅ No t.Skip() placeholders - actual implementations

**Advantage:** Uses your existing Cursor/Claude subscription - no additional API keys or costs!

## Example Workflow

```
1. analyze_repository → Get function-level map
2. get_test_recommendations → Get top 10 uncovered functions
3. generate_unit_tests → Generate tests for selected functions (uses Claude API)
4. verify_coverage_improvement → Verify coverage increased
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


# ContextForge

Go test coverage intelligence for your MCP-powered editor. ContextForge scans Go repositories, identifies test coverage gaps, and generates real table-driven tests — all from within Cursor, Claude Desktop, or Claude Code.

## What It Does

1. **Scans** your Go codebase — extracts every function, method, signature, and complexity score
2. **Generates scenarios** — determines which tests should exist (happy path, error cases, edge cases, boundary conditions)
3. **Finds gaps** — compares scenarios against your existing tests and reports what's missing
4. **Writes real tests** — generates compilable table-driven tests with assertions, not empty stubs

## Installation

Requires Go 1.23+.

```bash
go install github.com/diveshsaini1991/contextforge-2210991527/cmd/contextforge@latest
```

This installs the `contextforge` binary to your `$GOPATH/bin` directory.

Find your binary path:
```bash
# macOS / Linux
ls ~/go/bin/contextforge

# Windows
dir %USERPROFILE%\go\bin\contextforge.exe
```

### Cursor

Add to `.cursor/mcp.json` in your project (or globally at `~/.cursor/mcp.json`):

**macOS / Linux:**
```json
{
  "mcpServers": {
    "contextforge": {
      "command": "/Users/<your-username>/go/bin/contextforge"
    }
  }
}
```

**Windows:**
```json
{
  "mcpServers": {
    "contextforge": {
      "command": "C:\\Users\\<your-username>\\go\\bin\\contextforge.exe"
    }
  }
}
```

### Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows):

**macOS:**
```json
{
  "mcpServers": {
    "contextforge": {
      "command": "/Users/<your-username>/go/bin/contextforge"
    }
  }
}
```

**Windows:**
```json
{
  "mcpServers": {
    "contextforge": {
      "command": "C:\\Users\\<your-username>\\go\\bin\\contextforge.exe"
    }
  }
}
```

### Claude Code

```bash
claude mcp add contextforge ~/go/bin/contextforge
```

### From Source

```bash
git clone https://github.com/diveshsaini1991/contextforge-2210991527.git
cd contextforge-2210991527
go install ./cmd/contextforge
```

## Usage

Once connected, you get 7 tools available in your editor. Use them in order:

### Step 1: Build Context

```
Use build_repo_context with repo_path="/path/to/your/go/project"
```

Scans everything and saves to `.contextforge/context.json`.

### Step 2: Analyze Scenarios

```
Use analyze_test_scenarios with repo_path="/path/to/your/go/project"
```

Generates all test scenarios that should exist. Saves to `.contextforge/scenarios.json`.

### Step 3: Check Coverage

```
Use check_test_coverage with repo_path="/path/to/your/go/project"
```

Compares scenarios against existing tests. Shows what's covered and what's missing.

### Step 4: Generate Tests

```
Use generate_test_stubs with repo_path="/path/to/your/go/project"
```

Generates real table-driven tests for testable functions. For example, `func Multiply(a, b int) int` produces:

```go
func TestMultiply(t *testing.T) {
    tests := []struct {
        name string
        a    int
        b    int
        want int
    }{
        {"basic case", 1, 1, 1},
        {"another case", 5, 5, 25},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Multiply(tt.a, tt.b)
            if got != tt.want {
                t.Errorf("Multiply() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

Functions requiring mocks (HTTP handlers, gin handlers) get TODO comments with setup hints instead.

### Analyze a Remote Repo

Pass `repo_url` instead of `repo_path` to analyze any public Git repository:

```
Use build_repo_context with repo_url="https://github.com/someone/their-repo"
```

The repo is cloned to a temp directory and cached for subsequent tool calls.

### Retrieve Cached Results

- `get_context` — retrieve saved context without re-scanning
- `get_scenarios` — retrieve saved scenarios without re-analyzing
- `get_coverage_report` — retrieve saved coverage report

## Tool Parameters

| Parameter | Available On | Description |
|-----------|-------------|-------------|
| `repo_path` | All tools | Absolute path to a local Go repository |
| `repo_url` | All tools | Git URL to clone and analyze (alternative to repo_path) |
| `force` | Steps 1-3 | Skip cache and force fresh analysis |
| `package_filter` | Steps 1-3 | Only process packages matching this substring |

## How Scenario Generation Works

ContextForge generates test scenarios based on function characteristics:

| Scenario Type | When Generated |
|---------------|----------------|
| **Happy path** | Every exported function, or unexported with complexity > 1 |
| **Error case** | Functions whose signature returns `error` |
| **Edge case** | Exported functions with complexity > 5 and > 10 lines |
| **Boundary** | Exported functions with complexity > 2 and numeric/collection params |

## How Test Generation Works

ContextForge reads function signatures via Go AST and generates real tests:

- **Simple functions** (`int`, `string`, `bool`, `float` params/returns) — full table-driven tests with assertions
- **Error-returning functions** — `wantErr bool` pattern with both success and error cases
- **Slice/map returns** — uses `reflect.DeepEqual` for comparison
- **Methods** — auto-constructs the receiver (`recv := &Type{}`)
- **`context.Context` params** — injects `context.Background()` automatically
- **HTTP/gin handlers** — TODO comment with setup hints (requires mocking)
- **`main`/`init`** — skipped with explanation

## Project Structure

```
contextforge-2210991527/
├── cmd/contextforge/     # Entry point (stdio MCP server)
├── mcp/                  # MCP tool registrations and handlers
├── internal/
│   ├── builder/          # Context building, scenario analysis, coverage checking, test generation
│   ├── scanner/          # Go AST parsing and function extraction
│   └── models/           # Data structures
```

## Requirements

- Go 1.23+
- Target repository must have a `go.mod` file
- Git (only needed if using `repo_url` for remote repos)
- `goimports` (optional, for auto-fixing imports in generated tests)


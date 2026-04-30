# ContextForge

Go test coverage intelligence for your MCP-powered editor. ContextForge scans Go repositories, identifies test coverage gaps, and generates test stubs — all from within Cursor or Claude Desktop.

## What It Does

1. **Scans** your Go codebase — extracts every function, method, signature, and complexity score
2. **Generates scenarios** — determines which tests should exist (happy path, error cases, edge cases, boundary conditions)
3. **Finds gaps** — compares scenarios against your existing tests and reports what's missing
4. **Creates stubs** — generates test files with skipped stub functions ready for implementation

## Installation

Requires Go 1.23+.

```bash
go install github.com/divesh/contextforge/cmd/mcp-server@latest
```

This installs the `mcp-server` binary to your `$GOPATH/bin`.

### Cursor

Add to `.cursor/mcp.json` in your project (or globally at `~/.cursor/mcp.json`):

```json
{
  "mcpServers": {
    "contextforge": {
      "command": "mcp-server"
    }
  }
}
```

### Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows):

```json
{
  "mcpServers": {
    "contextforge": {
      "command": "mcp-server"
    }
  }
}
```

### Claude Code

```bash
claude mcp add contextforge mcp-server
```

### From Source

```bash
git clone https://github.com/divesh/contextforge.git
cd contextforge
go build -o mcp-server ./cmd/mcp-server
```

Then point your MCP config to the built binary path.

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

### Step 4: Generate Stubs

```
Use generate_test_stubs with repo_path="/path/to/your/go/project"
```

Creates test files with stub functions for every missing test. Each stub calls `t.Skip()` so your test suite still passes.

### Analyze a Remote Repo

You can also pass a `repo_url` instead of `repo_path` to analyze any public Git repository:

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
| **Happy path** | Every function |
| **Error case** | Functions returning `error` or with pointer receivers |
| **Edge case** | Exported functions with cyclomatic complexity > 5 |
| **Boundary** | Functions with numeric, slice, or map parameters |

## Project Structure

```
contextforge/
├── cmd/mcp-server/       # Entry point (stdio MCP server)
├── mcp/                  # MCP tool registrations and handlers
├── internal/
│   ├── builder/          # Context building, scenario analysis, coverage checking, stub generation
│   ├── scanner/          # Go AST parsing and function extraction
│   └── models/           # Data structures
└── testdata/             # Sample project for testing
```

## Requirements

- Go 1.23+
- Target repository must have a `go.mod` file
- Git (only needed if using `repo_url` for remote repos)

## License

MIT

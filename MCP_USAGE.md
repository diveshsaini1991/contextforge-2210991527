# ContextForge MCP — Quick Reference

## Tools

| Tool | Step | Description |
|------|------|-------------|
| `build_repo_context` | 1 | Scan repo, extract all functions and test mappings |
| `analyze_test_scenarios` | 2 | Generate test scenarios (happy, error, edge, boundary) |
| `check_test_coverage` | 3 | Compare scenarios vs existing tests, find gaps |
| `generate_test_stubs` | 4 | Create stub test files for missing tests |
| `get_context` | — | Retrieve cached context |
| `get_scenarios` | — | Retrieve cached scenarios |
| `get_coverage_report` | — | Retrieve cached coverage report |

## Typical Workflow

```
1. build_repo_context(repo_path="/path/to/repo")
2. analyze_test_scenarios(repo_path="/path/to/repo")
3. check_test_coverage(repo_path="/path/to/repo")
4. generate_test_stubs(repo_path="/path/to/repo")
```

Steps 2-3 auto-build context if not cached. Use `force=true` to rebuild from scratch.

## Remote Repos

Pass `repo_url` instead of `repo_path` to clone and analyze any public repo:

```
build_repo_context(repo_url="https://github.com/user/repo")
```

Cloned repos are cached — subsequent calls reuse the clone.

## Filtering

Use `package_filter` to scope analysis to specific packages:

```
build_repo_context(repo_path="/path/to/repo", package_filter="pkg/api")
```

## Cached Files

All results are saved under `.contextforge/` in the repo:

```
.contextforge/
├── context.json          # Function map and test mappings
├── scenarios.json        # Generated test scenarios
└── coverage_report.json  # Gap analysis report
```

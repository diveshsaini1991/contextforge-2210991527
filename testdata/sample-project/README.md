# Sample Calculator API

This is a sample Go project with **partial test coverage** designed to test ContextForge's capabilities.

## Project Structure

```
sample-project/
├── main.go                      # API server entry point
├── pkg/
│   ├── calculator/
│   │   ├── calculator.go        # Core calculation functions
│   │   └── calculator_test.go   # PARTIAL tests (only Add and Subtract tested)
│   └── api/
│       └── handlers.go          # HTTP handlers (NO TESTS)
└── internal/
    └── utils/
        ├── helpers.go           # Utility functions
        └── helpers_test.go      # PARTIAL tests (only IsEven tested)
```

## Coverage Gaps

This project intentionally has incomplete test coverage:

### pkg/calculator/calculator.go
- ✅ **Tested**: Add, Subtract
- ❌ **NOT Tested**: Multiply, Divide, Power, Factorial, PrimeFactors, isPrime

### pkg/api/handlers.go
- ❌ **NO TESTS**: All handlers (AddHandler, SubtractHandler, DivideHandler, FactorialHandler, SetupRoutes)

### internal/utils/helpers.go
- ✅ **Tested**: IsEven
- ❌ **NOT Tested**: FormatNumber, IsOdd, Max, Min, Abs, validateInput

## Test Coverage

Run coverage analysis:
```bash
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

Expected partial coverage:
- `pkg/calculator`: ~30-40% (only Add and Subtract tested)
- `pkg/api`: 0% (no tests)
- `internal/utils`: ~15-20% (only IsEven tested)

## Usage

ContextForge should be able to:
1. Scan this repository and identify all functions
2. Run coverage analysis and detect the gaps listed above
3. Prioritize uncovered functions (exported functions like Divide, Factorial should rank higher)
4. Generate tests for the uncovered functions
5. Verify coverage improvements after test generation

## API Endpoints

- `POST /api/v1/add` - Add two numbers
- `POST /api/v1/subtract` - Subtract two numbers
- `POST /api/v1/divide` - Divide two numbers
- `GET /api/v1/factorial/:n` - Calculate factorial of n
- `GET /health` - Health check

## Running the API

```bash
go run main.go
```

Test endpoints:
```bash
curl -X POST http://localhost:8080/api/v1/add -H "Content-Type: application/json" -d '{"a":5,"b":3}'
curl http://localhost:8080/api/v1/factorial/5
```

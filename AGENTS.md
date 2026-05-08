# Calorie Tracker - Agent Guide

This document provides the context and conventions needed for AI agents working on this project.

## Project Overview

A Go CLI application for tracking daily nutrition and water intake. Uses AI (LLM via NVIDIA NIM API) for natural language food parsing and nutritional analysis.

## Build & Run

```bash
# Install dependencies
go mod tidy

# Run the application
go run main.go

# Build binary
go build -o calorie-tracker main.go
```

## Test Commands

```bash
# Run all tests
go test ./...

# Run with coverage report
go test ./... -coverprofile=coverage.out -covermode=atomic
go tool cover -func=coverage.out

# Run specific package
go test ./services -v
go test ./db -v
```

## Quality Gates (5 Dimensions)

| # | Dimension | Tool | Threshold | Enforced In |
|---|-----------|------|-----------|-------------|
| 1 | **Test Coverage** | `go test -cover` | ≥ 80% | CI + Pre-commit |
| 2 | **Cyclomatic Complexity** | `gocyclo` | ≤ 20 | CI + Pre-commit |
| 3 | **Mutation Testing** | Custom bash script | ≥ 60% | CI only |
| 4 | **File Size** | `wc -l` | ≤ 500 lines | CI + Pre-commit |
| 5 | **Linting** | `golangci-lint` | Zero issues | CI + Pre-commit |

## Linting Rules (`.golangci.yml`)

Enabled linters:
- `govet` - Go vet analysis
- `staticcheck` - Static analysis
- `unused` - Unused code detection
- `ineffassign` - Ineffective assignments
- `gocyclo` - Cyclomatic complexity
- `gosimple` - Simplification suggestions
- `errcheck` - Unchecked errors
- `deadcode` - Dead code detection
- `structcheck` - Unused struct fields
- `varcheck` - Unused variables
- `typecheck` - Type checking
- `goimports` - Import formatting

## Architecture

### Package Structure

```
calorie-tracker/
├── commands/       # Cobra CLI commands
│   ├── add.go      # Add food entry
│   ├── review.go   # Weekly review
│   ├── report.go   # Generate reports
│   ├── root.go     # Root command
│   └── water.go    # Add water entry
├── config/         # Configuration (env vars, .env file)
├── data/           # Embedded JSON data (rules, etc.)
├── db/             # Database abstraction
│   ├── interfaces.go   # DBProvider interface
│   ├── sqlite.go       # SQLite implementation
│   └── mock_db.go      # Mock for testing
├── models/         # Data structures (Entry, Macros, etc.)
├── services/       # Business logic
│   ├── llm.go              # LLM service (NVIDIA NIM API)
│   ├── food_parser.go      # NLP food parsing
│   ├── nutrition_engine.go # Hybrid nutrition lookup
│   ├── tracker.go          # Tracker service
│   └── ...
├── tui/            # Bubble Tea TUI
│   ├── model.go      # Model definition
│   ├── update.go     # Update logic
│   └── view.go       # View rendering
└── utils/          # Time utilities
```

### Dependency Rules

- `commands/` → `services/`, `db/`, `config/`
- `services/` → `models/`, `db/`, `config/`
- `tui/` → `models/`, `services/`
- `db/` → `models/`
- `models/` → (no internal deps)
- **No circular dependencies allowed**

## Coding Conventions

### Go Style

- Follow standard Go conventions (`gofmt`, `goimports`)
- Use `golangci-lint` for additional checks
- Keep functions under 20 CCN
- Keep files under 500 lines
- Add tests for all new code

### Error Handling

- Always check errors
- Wrap errors with context: `fmt.Errorf("context: %w", err)`
- Return early on errors

### Testing

- Use table-driven tests
- Mock external dependencies (HTTP, DB)
- Name tests descriptively: `Test<Component>_<Scenario>_<ExpectedResult>`
- Aim for 80%+ coverage

## LLM Configuration

Default models (configurable via env vars):
- **Food Model**: `meta/llama-3.3-70b-instruct` (NVIDIA NIM)
- **Review Model**: `z-ai/glm-5.1` (NVIDIA NIM)

API: NVIDIA NIM (`https://integrate.api.nvidia.com/v1`)
Auth: `NVIDIA_API_KEY` env var

## Common Tasks

### Adding a New Command

1. Create `commands/<name>.go`
2. Implement Cobra command
3. Register in `commands/root.go`
4. Add tests

### Adding a New Service

1. Create `services/<name>.go`
2. Define interface in `services/interfaces.go` if needed
3. Implement service
4. Add tests
5. Wire into `TrackerService` if needed

### Database Changes

1. Update `db/interfaces.go` if schema changes
2. Update `db/sqlite.go` implementation
3. Add migration logic in `migrateExistingTables()`
4. Update `db/mock_db.go`
5. Add tests

## CI/CD

GitHub Actions workflow (`.github/workflows/ci.yml`):
- **test**: Run tests, check 80% coverage
- **complexity**: Check CCN ≤ 20
- **lint**: Run `golangci-lint`
- **mutation**: Run mutation testing (≥ 60%)
- **build**: Build binary, verify architecture

# Agent Instructions

- If the repository has a pre-commit hook, never bypass it (`--no-verify`, disabling hooks, etc.).
- Run the hook normally and carefully read the errors.
- Fix the underlying issues instead of skipping validation.
- Keep iterating until the pre-commit checks pass successfully.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `NVIDIA_API_KEY` | (required) | NVIDIA NIM API key |
| `OPENAI_BASE_URL` | `https://integrate.api.nvidia.com/v1` | LLM API base URL |
| `OPENAI_MODEL` | `meta/llama-3.3-70b-instruct` | Food parsing model |
| `OPENAI_MODEL2` | `z-ai/glm-5.1` | Review analysis model |
| `SERPAPI_KEY` | (optional) | SerpAPI key for nutrition lookup |
| `NUTRITION_PRIORITY` | `serpapi,fatsecret` | Nutrition provider priority |

# Calorie Tracker CLI

[![CI](https://github.com/Tatarotus/calorie-tracker/actions/workflows/ci.yml/badge.svg)](https://github.com/Tatarotus/calorie-tracker/actions/workflows/ci.yml)
[![Coverage](https://codecov.io/gh/Tatarotus/calorie-tracker/branch/main/graph/badge.svg)](https://codecov.io/gh/Tatarotus/calorie-tracker)
[![Go Version](https://img.shields.io/badge/go-1.26-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

A smart, interactive CLI tool to track your daily nutrition and water intake using AI-powered natural language processing and a persistent local database.

## Features

- **Natural Language Input**: Just type what you ate (e.g., "2 eggs and a coffee" or "100g de arroz") and let the AI estimate the macros.
- **Smart Cache Matching**: Prioritizes a local nutritional database (`food_cache`). It can automatically calculate values for different portions (e.g., if you have "100g of rice" cached, it can calculate for "250g").
- **Portuguese Support**: Robust normalization for Portuguese accents (feijão/feijao) and plurals (ovo/ovos fritos).
- **Daily Dashboard**: Real-time tracking of calories, protein, carbs, fat, and water.
- **Goal Setting**: Set personal goals and get AI-powered weekly reviews on your progress.
- **Interactive TUI**: Built with Bubble Tea for a smooth terminal experience.

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/Tatarotus/calorie-tracker.git
   cd calorie-tracker
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Set your Configuration (via Environment Variables or `.env` file):
   Create a `.env` file in the root directory (you can copy `.env.example` if available):
   ```env
   NVIDIA_API_KEY="your-nvidia-nim-api-key"
   SERPAPI_KEY="your-serpapi-key-optional"
   FATSECRET_CLIENT_ID="your-fatsecret-client-id-optional"
   FATSECRET_CLIENT_SECRET="your-fatsecret-client-secret-optional"
   ```
   Or export them directly:
   ```bash
   export NVIDIA_API_KEY="your-nvidia-nim-api-key"
   ```

4. Run the application:
   ```bash
   go run main.go
   ```

## Tech Stack

- **Go**: Language
- **Bubble Tea**: TUI Framework
- **SQLite**: Local storage
- **NVIDIA NIM API**: LLM API for nutritional analysis (default: `meta/llama-3.3-70b-instruct`)

## Development

### Prerequisites

- Go 1.26+
- `gocyclo` for cyclomatic complexity checks
- `golangci-lint` for linting
- `go-mutesting` for mutation testing (CI only)

### Quality Gates

This project enforces strict quality standards across 5 dimensions:

| Dimension | Tool | Threshold |
|-----------|------|-----------|
| Test Coverage | `go test -cover` | **≥ 80%** |
| Cyclomatic Complexity | `gocyclo` | **≤ 20** |
| Mutation Score | Custom script | **≥ 60%** |
| File Size | `wc -l` | **≤ 500 lines** |
| Linting | `golangci-lint` | Zero issues |

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -coverprofile=coverage.out -covermode=atomic
go tool cover -func=coverage.out

# Run specific package tests
go test ./services -v
go test ./db -v
```

### Running Quality Checks Locally

```bash
# Run all pre-commit checks (tests, coverage, CCN, linting, file size)
bash scripts/pre_commit_check.sh

# Run individual checks
bash scripts/ccn/check_ccn.sh          # Cyclomatic complexity
golangci-lint run ./...                # Linting
```

### Project Structure

```
calorie-tracker/
├── commands/       # CLI commands (add, review, water, etc.)
├── config/         # Configuration loading
├── data/           # Embedded reference data
├── db/             # Database layer (SQLite)
├── docs/           # Documentation and reports
├── models/         # Data models
├── scripts/        # Quality check scripts
├── services/       # Business logic (parsing, LLM, nutrition engine)
├── tests/          # Integration tests
├── tui/            # Terminal UI (Bubble Tea)
├── utils/          # Utility functions
├── main.go         # Entry point
└── go.mod          # Go module definition
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run quality checks (`bash scripts/pre_commit_check.sh`)
5. Commit your changes (`git commit -m 'feat: add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

All PRs must pass the CI pipeline: tests, 80% coverage, CCN ≤ 20, linting, and mutation score ≥ 60%.

## License

MIT

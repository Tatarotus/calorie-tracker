# Calorie Tracker CLI

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
   git clone https://github.com/your-username/calorie-tracker.git
   cd calorie-tracker
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Set your SambaNova API Key:
   ```bash
   export SAMBA_API_KEY="your-api-key-here"
   ```

4. Run the application:
   ```bash
   go run main.go
   ```

## Tech Stack

- **Go**: Language
- **Bubble Tea**: TUI Framework
- **SQLite**: Local storage
- **SambaNova Cloud**: LLM API for nutritional analysis

## License

MIT

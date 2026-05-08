# Calorie Tracker CLI - Usage Guide

The Calorie Tracker CLI provides both an interactive Terminal User Interface (TUI) and direct command-line execution for tracking your nutrition.

## Interactive Mode (TUI)

Simply run the application without arguments to launch the dashboard:

```bash
calorie-tracker
```

### Keyboard Shortcuts
- `Enter` or `Ctrl+m`: Add food (opens the input prompt)
- `w`: Add water (in ml)
- `g`: Set a new goal
- `t`: View today's log
- `k`: View this week's log
- `m`: View this month's log
- `r`: Run AI progress review
- `u`: Undo last entry
- `q`, `Ctrl+c`, or `Esc`: Quit the application

## Command-Line Mode

You can also bypass the interactive dashboard by passing specific commands:

### Add Food
Provide a natural language description of what you ate. The AI will parse it and estimate macros.
```bash
calorie-tracker add "2 scrambled eggs and a slice of toast"
calorie-tracker add "100g de arroz branco e 150g de peito de frango"
```

### Add Water
Log water intake in milliliters.
```bash
calorie-tracker water 500
```

### View Daily Report
Quickly see your stats for the current day without entering the TUI.
```bash
calorie-tracker report
```

### Run Review
Trigger an AI-powered review of your recent progress against your goals.
```bash
calorie-tracker review
```

## Configuration Options

You can customize the behavior by editing the `.env` file or exporting environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `NVIDIA_API_KEY` | Required API key for NVIDIA NIM | |
| `OPENAI_BASE_URL` | Base URL for LLM integration | `https://integrate.api.nvidia.com/v1` |
| `OPENAI_MODEL` | Model used for parsing food | `meta/llama-3.3-70b-instruct` |
| `OPENAI_MODEL2` | Model used for weekly reviews | `z-ai/glm-5.1` |
| `NUTRITION_PRIORITY` | Provider priority (comma separated) | `serpapi,fatsecret` |
| `FATSECRET_CLIENT_ID` | FatSecret Provider Client ID | |
| `FATSECRET_CLIENT_SECRET` | FatSecret Provider Secret | |
| `SERPAPI_KEY` | SerpAPI Provider Key | |

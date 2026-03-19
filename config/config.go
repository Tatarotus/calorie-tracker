package config

import (
	"os"
)

type Config struct {
	SambaAPIKey    string
	OpenAIBaseURL  string
	FoodModel      string
	ReviewModel    string
}

func Load() *Config {
	return &Config{
		SambaAPIKey:   getEnv("SAMBA_API_KEY", ""),
		OpenAIBaseURL: getEnv("OPENAI_BASE_URL", "https://api.sambanova.ai/v1"),
		FoodModel:     getEnv("OPENAI_MODEL", "Meta-Llama-3.1-8B-Instruct"),
		ReviewModel:   getEnv("OPENAI_MODEL2", "DeepSeek-R1-0528"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

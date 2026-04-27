package config

import (
	"os"
)

type Config struct {
	SambaAPIKey   string
	OpenAIBaseURL string
	FoodModel     string
	ReviewModel   string
}

func Load() *Config {
	return &Config{
		SambaAPIKey:   getEnv("NVIDIA_API_KEY", ""),
		OpenAIBaseURL: getEnv("OPENAI_BASE_URL", "https://integrate.api.nvidia.com/v1"),
		FoodModel:     getEnv("OPENAI_MODEL", "meta/llama-3.1-70b-instruct"),
		ReviewModel:   getEnv("OPENAI_MODEL2", "z-ai/glm4.7"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

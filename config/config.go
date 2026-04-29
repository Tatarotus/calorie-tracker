package config

import (
	"bufio"
	"os"
	"strings"
)

type Config struct {
	SambaAPIKey              string
	OpenAIBaseURL            string
	FoodModel                string
	ReviewModel              string
	FatSecretClientID        string
	FatSecretClientSecret    string
	FatSecretScope           string
	FatSecretRegion          string
	FatSecretLanguage        string
	FatSecretUseLocalization bool
	FatSecretTokenURL        string
	FatSecretAPIURL          string
}

func Load() *Config {
	loadDotEnv(".env")

	return &Config{
		SambaAPIKey:              getEnv("NVIDIA_API_KEY", ""),
		OpenAIBaseURL:            getEnv("OPENAI_BASE_URL", "https://integrate.api.nvidia.com/v1"),
		FoodModel:                getEnv("OPENAI_MODEL", "nvidia/nemotron-3-nano-omni-30b-a3b-reasoning"),
		ReviewModel:              getEnv("OPENAI_MODEL2", "z-ai/glm-5.1"),
		FatSecretClientID:        getEnv("FATSECRET_CLIENT_ID", ""),
		FatSecretClientSecret:    getEnv("FATSECRET_CLIENT_SECRET", ""),
		FatSecretScope:           getEnv("FATSECRET_SCOPE", "basic"),
		FatSecretRegion:          getEnv("FATSECRET_REGION", "BR"),
		FatSecretLanguage:        getEnv("FATSECRET_LANGUAGE", "pt"),
		FatSecretUseLocalization: getEnv("FATSECRET_USE_LOCALIZATION", "") == "true",
		FatSecretTokenURL:        getEnv("FATSECRET_TOKEN_URL", "https://oauth.fatsecret.com/connect/token"),
		FatSecretAPIURL:          getEnv("FATSECRET_API_URL", "https://platform.fatsecret.com/rest/server.api"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func loadDotEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		_ = os.Setenv(key, value)
	}
}

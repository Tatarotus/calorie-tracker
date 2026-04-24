package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Save original env vars
	origNVIDIA := os.Getenv("NVIDIA_API_KEY")
	origOPENAI_BASE := os.Getenv("OPENAI_BASE_URL")
	origOPENAI_MODEL := os.Getenv("OPENAI_MODEL")
	origOPENAI_MODEL2 := os.Getenv("OPENAI_MODEL2")
	defer func() {
		os.Setenv("NVIDIA_API_KEY", origNVIDIA)
		os.Setenv("OPENAI_BASE_URL", origOPENAI_BASE)
		os.Setenv("OPENAI_MODEL", origOPENAI_MODEL)
		os.Setenv("OPENAI_MODEL2", origOPENAI_MODEL2)
	}()

	// Clear env vars to test defaults
	os.Unsetenv("NVIDIA_API_KEY")
	os.Unsetenv("OPENAI_BASE_URL")
	os.Unsetenv("OPENAI_MODEL")
	os.Unsetenv("OPENAI_MODEL2")

	cfg := Load()

	if cfg.OpenAIBaseURL != "https://integrate.api.nvidia.com/v1" {
		t.Errorf("Expected default OpenAIBaseURL, got %s", cfg.OpenAIBaseURL)
	}
	if cfg.FoodModel != "meta/llama-3.1-70b-instruct" {
		t.Errorf("Expected default FoodModel, got %s", cfg.FoodModel)
	}
	if cfg.ReviewModel != "z-ai/glm4.7" {
		t.Errorf("Expected default ReviewModel, got %s", cfg.ReviewModel)
	}
}

func TestLoadWithEnvVars(t *testing.T) {
	// Save original env vars
	origNVIDIA := os.Getenv("NVIDIA_API_KEY")
	origOPENAI_BASE := os.Getenv("OPENAI_BASE_URL")
	origOPENAI_MODEL := os.Getenv("OPENAI_MODEL")
	origOPENAI_MODEL2 := os.Getenv("OPENAI_MODEL2")
	defer func() {
		os.Setenv("NVIDIA_API_KEY", origNVIDIA)
		os.Setenv("OPENAI_BASE_URL", origOPENAI_BASE)
		os.Setenv("OPENAI_MODEL", origOPENAI_MODEL)
		os.Setenv("OPENAI_MODEL2", origOPENAI_MODEL2)
	}()

	// Set custom env vars
	os.Setenv("NVIDIA_API_KEY", "test-key-123")
	os.Setenv("OPENAI_BASE_URL", "https://custom.url/v1")
	os.Setenv("OPENAI_MODEL", "custom-model")
	os.Setenv("OPENAI_MODEL2", "custom-model-2")

	cfg := Load()

	if cfg.SambaAPIKey != "test-key-123" {
		t.Errorf("Expected SambaAPIKey to be test-key-123, got %s", cfg.SambaAPIKey)
	}
	if cfg.OpenAIBaseURL != "https://custom.url/v1" {
		t.Errorf("Expected custom OpenAIBaseURL, got %s", cfg.OpenAIBaseURL)
	}
	if cfg.FoodModel != "custom-model" {
		t.Errorf("Expected custom FoodModel, got %s", cfg.FoodModel)
	}
	if cfg.ReviewModel != "custom-model-2" {
		t.Errorf("Expected custom ReviewModel, got %s", cfg.ReviewModel)
	}
}

func TestGetEnv(t *testing.T) {
	// Save original env var
	orig := os.Getenv("TEST_GET_ENV_VAR")
	defer func() {
		if orig == "" {
			os.Unsetenv("TEST_GET_ENV_VAR")
		} else {
			os.Setenv("TEST_GET_ENV_VAR", orig)
		}
	}()

	// Test with existing env var
	os.Setenv("TEST_GET_ENV_VAR", "test-value")
	result := getEnv("TEST_GET_ENV_VAR", "fallback")
	if result != "test-value" {
		t.Errorf("Expected test-value, got %s", result)
	}

	// Test with non-existing env var (should use fallback)
	os.Unsetenv("TEST_GET_ENV_VAR")
	result = getEnv("TEST_GET_ENV_VAR", "fallback-value")
	if result != "fallback-value" {
		t.Errorf("Expected fallback-value, got %s", result)
	}
}

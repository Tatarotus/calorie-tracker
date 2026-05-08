package config

//nolint:errcheck // Test file - env var errors are not critical
import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	chdirForTest(t, t.TempDir())

	// Save original env vars
	origNVIDIA := os.Getenv("NVIDIA_API_KEY")
	origOPENAI_BASE := os.Getenv("OPENAI_BASE_URL")
	origOPENAI_MODEL := os.Getenv("OPENAI_MODEL")
	origOPENAI_MODEL2 := os.Getenv("OPENAI_MODEL2")
	origFatSecretClientID := os.Getenv("FATSECRET_CLIENT_ID")
	origFatSecretClientSecret := os.Getenv("FATSECRET_CLIENT_SECRET")
	origFatSecretScope := os.Getenv("FATSECRET_SCOPE")
	origFatSecretRegion := os.Getenv("FATSECRET_REGION")
	origFatSecretLanguage := os.Getenv("FATSECRET_LANGUAGE")
	origFatSecretUseLocalization := os.Getenv("FATSECRET_USE_LOCALIZATION")
	defer func() {
		_ = os.Setenv("NVIDIA_API_KEY", origNVIDIA)
		_ = os.Setenv("OPENAI_BASE_URL", origOPENAI_BASE)
		_ = os.Setenv("OPENAI_MODEL", origOPENAI_MODEL)
		_ = os.Setenv("OPENAI_MODEL2", origOPENAI_MODEL2)
		_ = os.Setenv("FATSECRET_CLIENT_ID", origFatSecretClientID)
		_ = os.Setenv("FATSECRET_CLIENT_SECRET", origFatSecretClientSecret)
		_ = os.Setenv("FATSECRET_SCOPE", origFatSecretScope)
		_ = os.Setenv("FATSECRET_REGION", origFatSecretRegion)
		_ = os.Setenv("FATSECRET_LANGUAGE", origFatSecretLanguage)
		_ = os.Setenv("FATSECRET_USE_LOCALIZATION", origFatSecretUseLocalization)
	}()

	// Clear env vars to test defaults
	_ = os.Unsetenv("NVIDIA_API_KEY")
	_ = os.Unsetenv("OPENAI_BASE_URL")
	_ = os.Unsetenv("OPENAI_MODEL")
	_ = os.Unsetenv("OPENAI_MODEL2")
	_ = os.Unsetenv("FATSECRET_CLIENT_ID")
	_ = os.Unsetenv("FATSECRET_CLIENT_SECRET")
	_ = os.Unsetenv("FATSECRET_SCOPE")
	_ = os.Unsetenv("FATSECRET_REGION")
	_ = os.Unsetenv("FATSECRET_LANGUAGE")
	_ = os.Unsetenv("FATSECRET_USE_LOCALIZATION")

	cfg := Load()

	if cfg.OpenAIBaseURL != "https://integrate.api.nvidia.com/v1" {
		t.Errorf("Expected default OpenAIBaseURL, got %s", cfg.OpenAIBaseURL)
	}
	if cfg.FoodModel != "meta/llama-3.3-70b-instruct" {
		t.Errorf("Expected default FoodModel, got %s", cfg.FoodModel)
	}
	if cfg.ReviewModel != "z-ai/glm-5.1" {
		t.Errorf("Expected default ReviewModel, got %s", cfg.ReviewModel)
	}
	if cfg.FatSecretScope != "basic" {
		t.Errorf("Expected default FatSecretScope, got %s", cfg.FatSecretScope)
	}
	if cfg.FatSecretRegion != "BR" {
		t.Errorf("Expected default FatSecretRegion, got %s", cfg.FatSecretRegion)
	}
	if cfg.FatSecretLanguage != "pt" {
		t.Errorf("Expected default FatSecretLanguage, got %s", cfg.FatSecretLanguage)
	}
	if cfg.FatSecretUseLocalization {
		t.Error("Expected FatSecretUseLocalization to default false")
	}
}

func TestLoadWithEnvVars(t *testing.T) {
	chdirForTest(t, t.TempDir())

	// Save original env vars
	origNVIDIA := os.Getenv("NVIDIA_API_KEY")
	origOPENAI_BASE := os.Getenv("OPENAI_BASE_URL")
	origOPENAI_MODEL := os.Getenv("OPENAI_MODEL")
	origOPENAI_MODEL2 := os.Getenv("OPENAI_MODEL2")
	origFatSecretClientID := os.Getenv("FATSECRET_CLIENT_ID")
	origFatSecretClientSecret := os.Getenv("FATSECRET_CLIENT_SECRET")
	origFatSecretScope := os.Getenv("FATSECRET_SCOPE")
	origFatSecretRegion := os.Getenv("FATSECRET_REGION")
	origFatSecretLanguage := os.Getenv("FATSECRET_LANGUAGE")
	origFatSecretUseLocalization := os.Getenv("FATSECRET_USE_LOCALIZATION")
	defer func() {
		_ = os.Setenv("NVIDIA_API_KEY", origNVIDIA)
		_ = os.Setenv("OPENAI_BASE_URL", origOPENAI_BASE)
		_ = os.Setenv("OPENAI_MODEL", origOPENAI_MODEL)
		_ = os.Setenv("OPENAI_MODEL2", origOPENAI_MODEL2)
		_ = os.Setenv("FATSECRET_CLIENT_ID", origFatSecretClientID)
		_ = os.Setenv("FATSECRET_CLIENT_SECRET", origFatSecretClientSecret)
		_ = os.Setenv("FATSECRET_SCOPE", origFatSecretScope)
		_ = os.Setenv("FATSECRET_REGION", origFatSecretRegion)
		_ = os.Setenv("FATSECRET_LANGUAGE", origFatSecretLanguage)
		_ = os.Setenv("FATSECRET_USE_LOCALIZATION", origFatSecretUseLocalization)
	}()

	// Clear env vars to test defaults
	_ = os.Unsetenv("NVIDIA_API_KEY")
	_ = os.Unsetenv("OPENAI_BASE_URL")
	_ = os.Unsetenv("OPENAI_MODEL")
	_ = os.Unsetenv("OPENAI_MODEL2")
	_ = os.Unsetenv("FATSECRET_CLIENT_ID")
	_ = os.Unsetenv("FATSECRET_CLIENT_SECRET")
	_ = os.Unsetenv("FATSECRET_SCOPE")
	_ = os.Setenv("NVIDIA_API_KEY", "test-key-123")
	_ = os.Setenv("OPENAI_BASE_URL", "https://custom.url/v1")
	_ = os.Setenv("OPENAI_MODEL", "custom-model")
	_ = os.Setenv("OPENAI_MODEL2", "custom-model-2")
	_ = os.Setenv("FATSECRET_CLIENT_ID", "fat-id")
	_ = os.Setenv("FATSECRET_CLIENT_SECRET", "fat-secret")
	_ = os.Setenv("FATSECRET_SCOPE", "basic localization")
	_ = os.Setenv("FATSECRET_REGION", "BR")
	_ = os.Setenv("FATSECRET_LANGUAGE", "pt")
	_ = os.Setenv("FATSECRET_USE_LOCALIZATION", "true")

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
	if cfg.FatSecretClientID != "fat-id" {
		t.Errorf("Expected FatSecretClientID to be fat-id, got %s", cfg.FatSecretClientID)
	}
	if cfg.FatSecretClientSecret != "fat-secret" {
		t.Errorf("Expected FatSecretClientSecret to be fat-secret, got %s", cfg.FatSecretClientSecret)
	}
	if cfg.FatSecretScope != "basic localization" {
		t.Errorf("Expected custom FatSecretScope, got %s", cfg.FatSecretScope)
	}
	if !cfg.FatSecretUseLocalization {
		t.Error("Expected FatSecretUseLocalization to be true")
	}
}

func TestGetEnv(t *testing.T) {
	// Save original env var
	orig := os.Getenv("TEST_GET_ENV_VAR")
	defer func() {
		if orig == "" {
			_ = os.Unsetenv("TEST_GET_ENV_VAR")
		} else {
			_ = os.Setenv("TEST_GET_ENV_VAR", orig)
		}
	}()

	// Test with existing env var
	_ = os.Setenv("TEST_GET_ENV_VAR", "test-value")
	result := getEnv("TEST_GET_ENV_VAR", "fallback")
	if result != "test-value" {
		t.Errorf("Expected test-value, got %s", result)
	}

	// Test with non-existing env var (should use fallback)
	_ = os.Unsetenv("TEST_GET_ENV_VAR")
	result = getEnv("TEST_GET_ENV_VAR", "fallback-value")
	if result != "fallback-value" {
		t.Errorf("Expected fallback-value, got %s", result)
	}
}

func TestLoadDotEnv(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	content := `
# comment
NVIDIA_API_KEY=from-dotenv
FATSECRET_CLIENT_ID="dotenv-fat-id"
FATSECRET_CLIENT_SECRET='dotenv-fat-secret'
FATSECRET_SCOPE=basic
`
	if err := os.WriteFile(envPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	keys := []string{
		"NVIDIA_API_KEY",
		"FATSECRET_CLIENT_ID",
		"FATSECRET_CLIENT_SECRET",
		"FATSECRET_SCOPE",
	}
	for _, key := range keys {
		_ = os.Unsetenv(key)
	}

	loadDotEnv(envPath)

	if got := os.Getenv("NVIDIA_API_KEY"); got != "from-dotenv" {
		t.Errorf("Expected NVIDIA_API_KEY from dotenv, got %q", got)
	}
	if got := os.Getenv("FATSECRET_CLIENT_ID"); got != "dotenv-fat-id" {
		t.Errorf("Expected FATSECRET_CLIENT_ID from dotenv, got %q", got)
	}
	if got := os.Getenv("FATSECRET_CLIENT_SECRET"); got != "dotenv-fat-secret" {
		t.Errorf("Expected FATSECRET_CLIENT_SECRET from dotenv, got %q", got)
	}
}

func TestLoadDotEnvDoesNotOverrideEnvironment(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("NVIDIA_API_KEY=from-dotenv\n"), 0600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("NVIDIA_API_KEY", "from-env")
	loadDotEnv(envPath)

	if got := os.Getenv("NVIDIA_API_KEY"); got != "from-env" {
		t.Errorf("Expected existing environment to win, got %q", got)
	}
}

func chdirForTest(t *testing.T, dir string) {
	t.Helper()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatal(err)
		}
	})
}

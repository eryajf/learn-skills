package config

import (
	"os"
	"testing"
)

func TestLoadFromEnv(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "test-key")
	os.Setenv("CNB_TOKEN", "test-token")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.LLM.APIKey != "test-key" {
		t.Errorf("Expected API key 'test-key', got '%s'", cfg.LLM.APIKey)
	}
	if cfg.CNB.Token != "test-token" {
		t.Errorf("Expected CNB token 'test-token', got '%s'", cfg.CNB.Token)
	}

	os.Unsetenv("OPENAI_API_KEY")
	os.Unsetenv("CNB_TOKEN")
}

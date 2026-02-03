package config

import (
	"fmt"
	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	LLM LLMConfig `mapstructure:"llm"`
	CNB CNBConfig `mapstructure:"cnb"`
}

// LLMConfig holds LLM client configuration
type LLMConfig struct {
	APIKey  string `mapstructure:"api_key"`
	BaseURL string `mapstructure:"base_url"`
	Model   string `mapstructure:"model"`
}

// CNBConfig holds CNB platform configuration
type CNBConfig struct {
	Token   string `mapstructure:"token"`
	MCPURL  string `mapstructure:"mcp_url"`
	APIBase string `mapstructure:"api_base"`
}

// Load reads configuration from file and environment
func Load() (*Config, error) {
	v := viper.New()

	// Config file settings
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("$HOME/.cnb-assistant")

	// Environment variable settings
	v.SetEnvPrefix("CNB_ASSISTANT")
	v.AutomaticEnv()

	// Bind specific environment variables
	v.BindEnv("llm.api_key", "OPENAI_API_KEY")
	v.BindEnv("llm.base_url", "OPENAI_BASE_URL")
	v.BindEnv("llm.model", "OPENAI_MODEL")
	v.BindEnv("cnb.token", "CNB_TOKEN")
	v.BindEnv("cnb.mcp_url", "CNB_MCP_URL")
	v.BindEnv("cnb.api_base", "CNB_API_BASE")

	// Set defaults
	v.SetDefault("llm.base_url", "https://api.openai.com/v1")
	v.SetDefault("llm.model", "gpt-4")
	v.SetDefault("cnb.mcp_url", "https://mcp.cnb.cool/sse")
	v.SetDefault("cnb.api_base", "https://api.cnb.cool")

	// Try to read config file (optional)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found is OK, we'll use env vars
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate required fields
	if cfg.LLM.APIKey == "" {
		return nil, fmt.Errorf("LLM API key is required (set OPENAI_API_KEY or llm.api_key in config)")
	}
	if cfg.CNB.Token == "" {
		return nil, fmt.Errorf("CNB token is required (set CNB_TOKEN or cnb.token in config)")
	}

	return &cfg, nil
}

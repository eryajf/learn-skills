package main

import (
	"fmt"
	"os"
	"path/filepath"

	"cnb.cool/znb/learn-skills/internal/cli"
	"cnb.cool/znb/learn-skills/internal/config"
	"cnb.cool/znb/learn-skills/internal/llm"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize LLM client
	llmClient, err := llm.NewClient(cfg.LLM.APIKey, cfg.LLM.BaseURL, cfg.LLM.Model)
	if err != nil {
		return fmt.Errorf("failed to create LLM client: %w", err)
	}

	// Load skill
	skillPath := filepath.Join("skills", "cnb-skill", "SKILL.md")
	skillContent, err := os.ReadFile(skillPath)
	if err != nil {
		return fmt.Errorf("failed to read skill file: %w", err)
	}

	// Create assistant
	assistant := cli.NewAssistant(llmClient, string(skillContent))
	if err := assistant.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize assistant: %w", err)
	}

	// Determine mode based on arguments
	args := os.Args[1:]
	if len(args) == 0 {
		// Interactive mode
		return cli.RunInteractive(assistant)
	} else {
		// One-shot mode
		return cli.RunOneShot(assistant, args)
	}
}

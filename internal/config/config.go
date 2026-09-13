package config

import (
	"time"

	"github.com/pehlicd/crd-wizard/internal/ai"
)

// Config holds all application configuration
type Config struct {
	// Global Settings
	Kubeconfig string
	Context    string
	LogFormat  string
	LogLevel   string

	// AI Configuration
	AI AIConfig

	// Search Configuration
	Search SearchConfig

	// Web Server Configuration
	Web WebConfig
}

type AIConfig struct {
	Enabled         bool
	Provider        string
	Model           string
	OllamaHost      string
	OllamaNumCtx    int
	OllamaKeepAlive string
	RequestTimeout  time.Duration
	EnableCache     bool
	GeminiAPIKey    string
}

type SearchConfig struct {
	Enabled      bool
	Provider     string
	GoogleAPIKey string
	GoogleCX     string
}

type WebConfig struct {
	Port string
}

// New returns a new Config with default values
func New() *Config {
	return &Config{
		LogFormat: "text",
		LogLevel:  "info",
		AI: AIConfig{
			Provider:       "ollama",
			Model:          "pehlicd/crd-wizard",
			OllamaHost:     "http://localhost:11434",
			RequestTimeout: 2 * time.Minute,
			EnableCache:    true,
		},
		Search: SearchConfig{
			Enabled:  true,
			Provider: "ddg",
		},
		Web: WebConfig{
			Port: "8080",
		},
	}
}

// ToAIConfig converts internal config to ai package config
func (c *AIConfig) ToAIConfig(searchConfig SearchConfig) ai.Config {
	return ai.Config{
		Provider:        ai.Provider(c.Provider),
		Model:           c.Model,
		OllamaHost:      c.OllamaHost,
		OllamaNumCtx:    c.OllamaNumCtx,
		OllamaKeepAlive: c.OllamaKeepAlive,
		RequestTimeout:  c.RequestTimeout,
		EnableCache:     c.EnableCache,
		GeminiAPIKey:    c.GeminiAPIKey,

		// Search
		EnableSearch:   searchConfig.Enabled,
		SearchProvider: ai.SearchProvider(searchConfig.Provider),
		GoogleAPIKey:   searchConfig.GoogleAPIKey,
		GoogleCX:       searchConfig.GoogleCX,
	}
}

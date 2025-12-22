package ai

import (
	"context"
	"time"
)

// LLMProvider defines the interface for interacting with Large Language Models.
type LLMProvider interface {
	// Generate sends a prompt to the LLM and returns the generated text.
	Generate(ctx context.Context, prompt string) (string, error)
	// Name returns the provider name.
	Name() string
}

type Provider string

const (
	ProviderOllama    Provider = "ollama"
	ProviderGemini    Provider = "gemini"
	ProviderOpenAI    Provider = "openai"
	ProviderAnthropic Provider = "anthropic"
)

type Config struct {
	Provider Provider
	Model    string

	// Generic Timeouts
	RequestTimeout time.Duration

	OllamaHost string

	// Performance Configuration
	OllamaNumCtx    int    // Context window size (e.g., 4096)
	OllamaKeepAlive string // Duration to keep model loaded (e.g., "5m")
	EnableCache     bool   // Toggle in-memory caching

	// Validation Configuration
	MaxValidationRetries int // How many times to retry if dry-run fails (suggest 3)

	// Search Configuration
	EnableSearch   bool
	SearchProvider SearchProvider // "google" or "ddg"
	GoogleAPIKey   string         // Only needed if Provider is "google"
	GoogleCX       string         // Only needed if Provider is "google"

	// Gemini Configuration
	GeminiAPIKey string
}

type SearchProvider string

const (
	SearchProviderGoogle     SearchProvider = "google"
	SearchProviderDuckDuckGo SearchProvider = "ddg"
)

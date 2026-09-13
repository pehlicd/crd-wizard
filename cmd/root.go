/*
Copyright © 2025 Furkan Pehlivan furkanpehlivan34@gmail.com

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"os"

	"time"

	"github.com/pehlicd/crd-wizard/internal/config"
	"github.com/spf13/cobra"
)

// cfg holds the global application configuration
var cfg = config.New()

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "crd-wizard",
	Short: "A tool to explore Kubernetes CRDs via a TUI or web interface.",
	Long: `crd-wizard is a powerful CLI application that provides two ways to
explore Custom Resource Definitions (CRDs) in your Kubernetes cluster:

- A beautiful and interactive Terminal User Interface (TUI)
- A simple web server providing a JSON API for CRDs`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Persistent Flags
	rootCmd.PersistentFlags().StringVar(&cfg.Kubeconfig, "kubeconfig", "", "path to the kubeconfig file (optional)")
	rootCmd.PersistentFlags().StringVar(&cfg.Context, "context", "", "context name (optional)")
	rootCmd.PersistentFlags().StringVar(&cfg.LogFormat, "log-format", "text", "log format")
	rootCmd.PersistentFlags().StringVar(&cfg.LogLevel, "log-level", "info", "log level")

	// AI Flags
	rootCmd.PersistentFlags().BoolVar(&cfg.AI.Enabled, "enable-ai", false, "Enable AI features")
	rootCmd.PersistentFlags().StringVar(&cfg.AI.Provider, "ai-provider", "ollama", "AI provider to use (ollama, gemini, etc.)")
	rootCmd.PersistentFlags().StringVar(&cfg.AI.Model, "ai-model", "pehlicd/crd-wizard", "Model to use for AI analysis and generation")
	rootCmd.PersistentFlags().StringVar(&cfg.AI.OllamaHost, "ollama-host", "http://localhost:11434", "Ollama API host (only for ollama provider)")
	rootCmd.PersistentFlags().IntVar(&cfg.AI.OllamaNumCtx, "ollama-num-ctx", 0, "Ollama context window size")
	rootCmd.PersistentFlags().StringVar(&cfg.AI.OllamaKeepAlive, "ollama-keep-alive", "", "Ollama keep-alive duration")

	// Helper to handle duration flag
	var timeout int
	rootCmd.PersistentFlags().IntVar(&timeout, "request-timeout", 2, "Timeout in minutes for AI requests")
	cobra.OnInitialize(func() {
		cfg.AI.RequestTimeout = time.Duration(timeout) * time.Minute
	})

	rootCmd.PersistentFlags().BoolVar(&cfg.AI.EnableCache, "enable-cache", true, "Enable caching of AI responses")

	// Search Flags
	rootCmd.PersistentFlags().BoolVar(&cfg.Search.Enabled, "enable-search", true, "Enable web search for CRD documentation (requires enable-ai)")
	rootCmd.PersistentFlags().StringVar(&cfg.Search.Provider, "search-provider", "ddg", "Search provider to use: 'ddg' (DuckDuckGo, free) or 'google' (Requires API Key)")
	rootCmd.PersistentFlags().StringVar(&cfg.Search.GoogleAPIKey, "google-api-key", "", "Google Custom Search API Key (required if search-provider is google)")
	rootCmd.PersistentFlags().StringVar(&cfg.Search.GoogleCX, "google-cx", "", "Google Custom Search Engine ID (required if search-provider is google)")

	rootCmd.PersistentFlags().StringVar(&cfg.AI.GeminiAPIKey, "gemini-api-key", "", "Gemini API Key (required if ai-provider is gemini)")
}

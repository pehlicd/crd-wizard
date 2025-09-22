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

	"github.com/pehlicd/crd-wizard/internal/ai"
	"github.com/pehlicd/crd-wizard/internal/k8s"
	"github.com/pehlicd/crd-wizard/internal/logger"
	"github.com/pehlicd/crd-wizard/internal/web"

	"github.com/spf13/cobra"
)

// Configuration variables bound to flags
var (
	port           string
	enableAI       bool
	ollamaHost     string
	ollamaModel    string
	requestTimeout int // in minutes
	numCtx         int
	keepAlive      string
	enableCache    bool

	// New Search Configuration Flags
	enableSearch   bool
	searchProvider string
	googleAPIKey   string
	googleCX       string
)

// webCmd represents the web command
var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Launch a web server to serve CRD data via a JSON API.",
	Long:  `The web server exposes endpoints to list CRDs, their instances, and related events. It can be used as a backend for a graphical user interface.`,
	Run: func(_ *cobra.Command, _ []string) {
		log := logger.NewLogger(logFormat, logLevel, os.Stderr)

		client, err := k8s.NewClient(kubeconfig, context, log)
		if err != nil {
			log.Error("unable to create k8s client", "err", err)
			os.Exit(1)
		}

		var aiClient *ai.Client

		if enableAI {
			// Construct the AI Config from flags
			aiConfig := ai.Config{
				OllamaHost:     ollamaHost,
				OllamaModel:    ollamaModel,
				RequestTimeout: time.Duration(requestTimeout) * time.Minute,
				NumCtx:         numCtx,
				KeepAlive:      keepAlive,
				EnableCache:    enableCache,

				// Search Configuration
				EnableSearch:   enableSearch,
				SearchProvider: ai.SearchProvider(searchProvider),
				GoogleAPIKey:   googleAPIKey,
				GoogleCX:       googleCX,
			}

			aiClient = ai.NewClient(aiConfig, client, log)

			log.Info("AI features enabled",
				"ollama_host", ollamaHost,
				"ollama_model", ollamaModel,
				"search_enabled", enableSearch,
				"search_provider", searchProvider,
			)
		}

		server := web.NewServer(client, port, aiClient, log)
		log.Info("starting web server", "port", port)
		if err := server.Start(); err != nil {
			log.Error("error starting web server", "err", err)
			os.Exit(1)
		}
	},
}

func init() {
	// Server Flags
	webCmd.Flags().StringVarP(&port, "port", "p", "8080", "Port for the web server")

	// AI Flags
	webCmd.Flags().BoolVar(&enableAI, "enable-ai", false, "Enable AI features via Ollama")
	webCmd.Flags().StringVar(&ollamaHost, "ollama-host", "http://localhost:11434", "Ollama API host")
	webCmd.Flags().StringVar(&ollamaModel, "ollama-model", "llama3.1", "Ollama model to use for generation")
	webCmd.Flags().IntVar(&requestTimeout, "request-timeout", 2, "Timeout in minutes for AI requests")
	webCmd.Flags().IntVar(&numCtx, "num-ctx", 0, "Number of context CRD examples to provide to the AI model")
	webCmd.Flags().StringVar(&keepAlive, "keep-alive", "", "Keep AI model instances alive between requests")
	webCmd.Flags().BoolVar(&enableCache, "enable-cache", true, "Enable caching of AI responses to improve performance")

	// Search Flags
	webCmd.Flags().BoolVar(&enableSearch, "enable-search", true, "Enable web search for CRD documentation (requires enable-ai)")
	webCmd.Flags().StringVar(&searchProvider, "search-provider", "ddg", "Search provider to use: 'ddg' (DuckDuckGo, free) or 'google' (Requires API Key)")
	webCmd.Flags().StringVar(&googleAPIKey, "google-api-key", "", "Google Custom Search API Key (required if search-provider is google)")
	webCmd.Flags().StringVar(&googleCX, "google-cx", "", "Google Custom Search Engine ID (required if search-provider is google)")

	rootCmd.AddCommand(webCmd)
}

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

	"github.com/pehlicd/crd-wizard/internal/ai"
	"github.com/pehlicd/crd-wizard/internal/k8s"
	"github.com/pehlicd/crd-wizard/internal/logger"
	"github.com/pehlicd/crd-wizard/internal/web"

	"github.com/spf13/cobra"
)

// webCmd represents the web command
var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Launch a web server to serve CRD data via a JSON API.",
	Long:  `The web server exposes endpoints to list CRDs, their instances, and related events. It can be used as a backend for a graphical user interface.`,
	Run: func(_ *cobra.Command, _ []string) {
		log := logger.NewLogger(cfg.LogFormat, cfg.LogLevel, os.Stderr)

		clusterManager, err := k8s.NewClusterManager(cfg.Kubeconfig, log)
		if err != nil {
			log.Error("unable to create cluster manager", "err", err)
			os.Exit(1)
		}

		var aiClient *ai.Client

		if cfg.AI.Enabled {
			// Construct the AI Config from flags
			aiConfig := cfg.AI.ToAIConfig(cfg.Search)

			// AI client needs a single K8s client for context fetching, use current
			aiClient = ai.NewClient(aiConfig, clusterManager.GetCurrentClient(), log)

			log.Info("AI features enabled",
				"provider", cfg.AI.Provider,
				"model", cfg.AI.Model,
				"ollama_host", cfg.AI.OllamaHost,
				"search_enabled", cfg.Search.Enabled,
				"search_provider", cfg.Search.Provider,
			)
		}

		server := web.NewServer(clusterManager, cfg.Web.Port, aiClient, log)
		log.Info("starting web server", "port", cfg.Web.Port, "clusters", clusterManager.ClusterCount())
		if err := server.Start(); err != nil {
			log.Error("error starting web server", "err", err)
			os.Exit(1)
		}
	},
}

func init() {
	// Server Flags
	webCmd.Flags().StringVarP(&cfg.Web.Port, "port", "p", "8080", "Port for the web server")

	rootCmd.AddCommand(webCmd)
}

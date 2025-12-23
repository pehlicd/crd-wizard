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
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pehlicd/crd-wizard/internal/ai"
	"github.com/pehlicd/crd-wizard/internal/k8s"
	"github.com/pehlicd/crd-wizard/internal/logger"
	"github.com/pehlicd/crd-wizard/internal/tui"

	"github.com/spf13/cobra"
)

var crd, kind string

// tuiCmd represents the tui command
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the interactive Terminal User Interface to explore CR(D)s.",
	Long: `Launch an interactive Terminal User Interface (TUI) to explore Custom
Resource Definitions (CRDs) and their corresponding Custom Resources (CRs)
in your cluster.

The TUI provides a rich experience for navigating CRDs, viewing their
instances, and inspecting definitions and events. You can optionally start
the TUI pre-focused on a specific CRD or Kind.`,
	Example: `
  # Launch the TUI and browse all CRDs
  crd-wizard tui

  # Launch and focus directly on a specific CRD
  crd-wizard tui --crd alertmanagers.monitoring.coreos.com

  # Launch and focus directly on a specific Kind
  crd-wizard tui --kind Alertmanager

  # Launch and focus on a Kind and specific CRD
  crd-wizard tui --crd alertmanagers.monitoring.coreos.com --kind Prometheus`,
	Run: func(_ *cobra.Command, _ []string) {
		log := logger.NewLogger(logFormat, logLevel, io.Discard)

		// Initialize the ClusterManager to load all contexts.
		clusterManager, err := k8s.NewClusterManager(kubeconfig, log)
		if err != nil {
			fmt.Printf("❌ Could not create cluster manager: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Loaded %d cluster(s) from kubeconfig\n", clusterManager.ClusterCount())

		var aiClient *ai.Client
		if enableAI {
			aiConfig := ai.Config{
				Provider:        ai.Provider(aiProvider),
				Model:           aiModel,
				OllamaHost:      ollamaHost,
				RequestTimeout:  time.Duration(requestTimeout) * time.Minute,
				OllamaNumCtx:    ollamaNumCtx,
				OllamaKeepAlive: ollamaKeepAlive,
				EnableCache:     enableCache,
				EnableSearch:    enableSearch,
				SearchProvider:  ai.SearchProvider(searchProvider),
				GoogleAPIKey:    googleAPIKey,
				GoogleCX:        googleCX,
				GeminiAPIKey:    geminiAPIKey,
			}
			// AI client needs a single K8s client for context fetching, use current
			aiClient = ai.NewClient(aiConfig, clusterManager.GetCurrentClient(), log)
		}

		// Start the TUI.
		if err := tui.Start(clusterManager, aiClient, crd, kind); err != nil {
			fmt.Printf("❌ TUI Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	tuiCmd.Flags().StringVar(&crd, "crd", "", "Focus on a specific Custom Resource Definition by name (e.g., 'alertmanagers.monitoring.coreos.com') (optional)")
	tuiCmd.Flags().StringVar(&kind, "kind", "", "Focus on a specific Kind (e.g., 'Prometheus') (optional)")
	rootCmd.AddCommand(tuiCmd)
}

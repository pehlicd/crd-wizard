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
	"os"

	"github.com/pehlicd/crd-wizard/internal/k8s"
	"github.com/pehlicd/crd-wizard/internal/tui"

	"github.com/spf13/cobra"
)

// tuiCmd represents the tui command
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the interactive Terminal User Interface to explore CR(D)s.",
	Long:  `The TUI provides a rich, interactive experience for navigating CRDs, viewing their instances, and inspecting their definitions and events.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize the Kubernetes client.
		client, err := k8s.NewClient(kubeconfig, context)
		if err != nil {
			fmt.Printf("❌ Could not create Kubernetes client: %v\n", err)
			os.Exit(1)
		}

		// Start the TUI.
		if err := tui.Start(client); err != nil {
			fmt.Printf("❌ TUI Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}

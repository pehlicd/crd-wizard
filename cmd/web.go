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
	"github.com/pehlicd/crd-wizard/internal/web"

	"github.com/spf13/cobra"
)

var port string

// webCmd represents the web command
var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Launch a web server to serve CRD data via a JSON API.",
	Long:  `The web server exposes endpoints to list CRDs, their instances, and related events. It can be used as a backend for a graphical user interface.`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := k8s.NewClient(kubeconfig, context)
		if err != nil {
			fmt.Printf("❌ Could not create Kubernetes client: %v\n", err)
			os.Exit(1)
		}

		server := web.NewServer(client)
		fmt.Printf("🚀 Starting web server on http://localhost:%s\n", port)
		if err := server.Start(port); err != nil {
			fmt.Printf("❌ Could not start web server: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	webCmd.Flags().StringVarP(&port, "port", "p", "8080", "Port for the web server")
	rootCmd.AddCommand(webCmd)
}

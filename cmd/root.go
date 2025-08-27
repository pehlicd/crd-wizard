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

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "crd-wizard",
	Short: "A tool to explore Kubernetes CRDs via a TUI or web interface.",
	Long: `crd-wizard is a powerful CLI application that provides two ways to
explore Custom Resource Definitions (CRDs) in your Kubernetes cluster:

- A beautiful and interactive Terminal User Interface (TUI)
- A simple web server providing a JSON API for CRDs`,
}

var (
	kubeconfig, context,
	logFormat, logLevel string
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "path to the kubeconfig file (optional)")
	rootCmd.PersistentFlags().StringVar(&context, "context", "", "context name (optional)")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "text", "log format")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level")
}

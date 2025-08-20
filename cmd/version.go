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

	"github.com/spf13/cobra"
)

var (
	versionString string
	buildDate     string
	buildCommit   string
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information of crd-explorer",
	Long:  `This command will print the version information of crd-explorer and exit.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("CR(D) Explorer version: %s\n", versionString)
		fmt.Printf("Build date: %s\n", buildDate)
		fmt.Printf("Build commit: %s\n", buildCommit)
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

/*
Copyright © 2025 Furkan Pehlivan furkanpehlivan34@gmail.com
*/
package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/pehlicd/crd-wizard/internal/generator"
	"github.com/pehlicd/crd-wizard/internal/k8s"
	"github.com/pehlicd/crd-wizard/internal/logger"
	"github.com/pehlicd/crd-wizard/internal/models"
	"github.com/spf13/cobra"
)

var (
	exportAll    bool
	exportFormat string
	exportOutput string
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export [crd-name]",
	Short: "Export documentation for CRDs from the cluster",
	Long: `Export documentation for Custom Resource Definitions (CRDs) present in the connected Kubernetes cluster.
You can export a single CRD by name or all CRDs using the --all flag.
Supported formats are HTML and Markdown.`,
	Example: `
  # Export a single CRD to HTML (default)
  crd-wizard export alertmanagers.monitoring.coreos.com

  # Export all CRDs to Markdown
  crd-wizard export --all --format md --output ./docs/

  # Export to specific file
  crd-wizard export prometheuses.monitoring.coreos.com -o prometheus.html
`,
	Run: func(cmd *cobra.Command, args []string) {
		log := logger.NewLogger(logFormat, logLevel, os.Stderr)

		if !exportAll && len(args) == 0 {
			log.Error("error: you must specify a CRD name or use --all")
			os.Exit(1)
		}

		client, err := k8s.NewClient(kubeconfig, context, log)
		if err != nil {
			log.Error("unable to create k8s client", "err", err)
			os.Exit(1)
		}

		gen := generator.NewGenerator()

		if exportAll {
			// List all CRDs
			crds, err := client.GetCRDs(cmd.Context()) // This returns models.CRD, not full CRD
			if err != nil {
				log.Error("failed to list CRDs", "err", err)
				os.Exit(1)
			}

			// We need full CRDs for generation.
			// GetCRDs returns a simplified struct. We might need to iterate and fetch full CRDs.
			// Or modify GetCRDs to return what we need, but that might affect other parts.
			// Better to just fetch names and then GetFullCRD for each.

			for _, simpleCRD := range crds {
				fullCRD, err := client.GetFullCRD(cmd.Context(), simpleCRD.Name)
				if err != nil {
					log.Error("failed to get full CRD", "name", simpleCRD.Name, "err", err)
					continue
				}

				// Convert to APICRD
				apiCRD := models.ToAPICRD(*fullCRD, 0)

				content, err := gen.Generate(apiCRD, exportFormat)
				if err != nil {
					log.Error("failed to generate documentation", "name", simpleCRD.Name, "err", err)
					continue
				}

				filename := fmt.Sprintf("%s.%s", simpleCRD.Name, getExtension(exportFormat))
				if exportOutput != "" {
					// Assume exportOutput is a directory for --all
					// Ensure directory exists
					_ = os.MkdirAll(exportOutput, 0755)
					filename = fmt.Sprintf("%s/%s", exportOutput, filename)
				}

				err = os.WriteFile(filename, content, 0644)
				if err != nil {
					log.Error("failed to write file", "file", filename, "err", err)
					continue
				}
				log.Info("generated documentation", "file", filename)
			}

		} else {
			crdName := args[0]
			fullCRD, err := client.GetFullCRD(cmd.Context(), crdName)
			if err != nil {
				log.Error("failed to get CRD", "name", crdName, "err", err)
				os.Exit(1)
			}

			apiCRD := models.ToAPICRD(*fullCRD, 0)
			content, err := gen.Generate(apiCRD, exportFormat)
			if err != nil {
				log.Error("failed to generate documentation", "err", err)
				os.Exit(1)
			}

			outputTarget := exportOutput
			if outputTarget == "" {
				outputTarget = fmt.Sprintf("%s.%s", crdName, getExtension(exportFormat))
			}

			// If output is '-', write to stdout
			if outputTarget == "-" {
				_, err = io.Writer(os.Stdout).Write(content)
				if err != nil {
					log.Error("failed to write to stdout", "err", err)
				}
			} else {
				err = os.WriteFile(outputTarget, content, 0644)
				if err != nil {
					log.Error("failed to write file", "file", outputTarget, "err", err)
					os.Exit(1)
				}
				log.Info("generated documentation", "file", outputTarget)
			}
		}
	},
}

func getExtension(format string) string {
	if format == "markdown" || format == "md" {
		return "md"
	}
	return "html"
}

func init() {
	exportCmd.Flags().BoolVar(&exportAll, "all", false, "Export all CRDs in the cluster")
	exportCmd.Flags().StringVar(&exportFormat, "format", "html", "Output format (html or markdown)")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output path (file or directory)")

	rootCmd.AddCommand(exportCmd)
}

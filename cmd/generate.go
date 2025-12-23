/*
Copyright © 2025 Furkan Pehlivan furkanpehlivan34@gmail.com
*/
package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/pehlicd/crd-wizard/internal/generator"
	"github.com/pehlicd/crd-wizard/internal/logger"
	"github.com/pehlicd/crd-wizard/internal/models"
	"github.com/pehlicd/crd-wizard/internal/util"
	"github.com/spf13/cobra"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

var (
	generateFile string
	generateUrl  string
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate documentation from a local CRD file",
	Long: `Generate documentation (HTML or Markdown) from a local CRD YAML or JSON file.
This allows you to verify documentation before applying the CRD to a cluster, or to generate docs in CI/CD pipelines.`,
	Example: `
  # Generate HTML from a local file
  crd-wizard generate -f ./crd.yaml

  # Generate Markdown and output to stdout
  crd-wizard generate -f ./crd.yaml --format md -o -
`,
	Run: func(cmd *cobra.Command, args []string) {
		log := logger.NewLogger(logFormat, logLevel, os.Stderr)

		if generateFile == "" && generateUrl == "" {
			log.Error("error: --file or --url flag is required")
			os.Exit(1)
		}

		var data []byte
		var err error

		if generateUrl != "" {
			rawURL := util.ConvertGitUrlToRaw(generateUrl)
			// log.Info("fetching CRD from URL", "original", generateUrl, "raw", rawURL) // Info log might pollute output if stdout is used for content? No, stdout is used for generated content. Logs go to Stderr.

			resp, err := http.Get(rawURL)
			if err != nil {
				log.Error("failed to fetch CRD from URL", "url", rawURL, "err", err)
				os.Exit(1)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				log.Error("failed to fetch CRD from URL", "url", rawURL, "status", resp.Status)
				os.Exit(1)
			}

			// Read limited amount to prevent abuse
			const maxFileSize = 10 * 1024 * 1024 // 10MB
			data, err = io.ReadAll(io.LimitReader(resp.Body, maxFileSize))
			if err != nil {
				log.Error("failed to read CRD content", "err", err)
				os.Exit(1)
			}
		} else {
			// Read file
			data, err = os.ReadFile(generateFile)
			if err != nil {
				log.Error("failed to read file", "file", generateFile, "err", err)
				os.Exit(1)
			}
		}

		// Parse YAML/JSON to CRD
		var crd apiextensionsv1.CustomResourceDefinition
		if err := yaml.Unmarshal(data, &crd); err != nil {
			log.Error("failed to parse CRD", "err", err)
			os.Exit(1)
		}

		gen := generator.NewGenerator()
		apiCRD := models.ToAPICRD(crd, 0)

		content, err := gen.Generate(apiCRD, exportFormat)
		if err != nil {
			log.Error("failed to generate documentation", "err", err)
			os.Exit(1)
		}

		outputTarget := exportOutput
		if outputTarget == "" {
			// auto-generate name based on file but change extension
			outputTarget = fmt.Sprintf("doc.%s", getExtension(exportFormat))
		}

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
	},
}

func init() {
	generateCmd.Flags().StringVarP(&generateFile, "file", "f", "", "Path to the CRD file (YAML or JSON)")
	generateCmd.Flags().StringVarP(&generateUrl, "url", "u", "", "URL to the CRD file (supports GitHub/GitLab blob URLs)")
	generateCmd.Flags().StringVar(&exportFormat, "format", "html", "Output format (html or markdown)")
	generateCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output path (file or directory, use - for stdout)")

	rootCmd.AddCommand(generateCmd)
}

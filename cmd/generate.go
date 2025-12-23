/*
Copyright © 2025 Furkan Pehlivan furkanpehlivan34@gmail.com
*/
package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/pehlicd/crd-wizard/internal/generator"
	"github.com/pehlicd/crd-wizard/internal/giturl"
	"github.com/pehlicd/crd-wizard/internal/logger"
	"github.com/pehlicd/crd-wizard/internal/models"
)

var (
	generateFile string
	generateURL  string
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate documentation from a CRD file",
	Long: `Generate documentation from a CRD file in either HTML or Markdown format.
Example:
  crd-wizard generate -f path/to/crd.yaml -o html > doc.html
  crd-wizard generate -f path/to/crd.yaml -o markdown > doc.md`,
	Run: func(_ *cobra.Command, _ []string) {
		log := logger.NewLogger(logFormat, logLevel, os.Stderr)

		var crdContent []byte
		var err error

		if generateURL != "" {
			rawURL := giturl.ConvertGitURLToRaw(generateURL)

			resp, err := http.Get(rawURL) //nolint:gosec // user supplied url is intended used for CLI purposes
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
			crdContent, err = io.ReadAll(io.LimitReader(resp.Body, maxFileSize))
			if err != nil {
				log.Error("failed to read CRD content", "err", err)
				os.Exit(1)
			}
		} else if generateFile != "" {
			// Read file
			crdContent, err = os.ReadFile(generateFile)
			if err != nil {
				log.Error("failed to read file", "file", generateFile, "err", err)
				os.Exit(1)
			}
		} else {
			log.Error("error: --file or --url flag is required")
			os.Exit(1)
		}

		// Parse YAML/JSON to CRD
		var crd apiextensionsv1.CustomResourceDefinition
		if err := yaml.Unmarshal(crdContent, &crd); err != nil {
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
			err = os.WriteFile(outputTarget, content, 0644) //nolint:gosec // 0644 is intended for documentation
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
	generateCmd.Flags().StringVarP(&generateURL, "url", "u", "", "URL to the CRD file (Git provider)")
	generateCmd.Flags().StringVar(&exportFormat, "format", "html", "Output format (html or markdown)")
	generateCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output path (file or directory, use - for stdout)")

	rootCmd.AddCommand(generateCmd)
}

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
package generator

import (
	"bytes"
	"fmt"
	"sort"
	"text/template"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/pehlicd/crd-wizard/internal/models"
)

// Generator handles the generation of documentation from CRDs.
type Generator struct{}

// NewGenerator creates a new Generator.
func NewGenerator() *Generator {
	return &Generator{}
}

// DocData represents the data structure passed to the templates.
type DocData struct {
	APIVersion   string
	Kind         string
	ResourceKind string
	Metadata     DocMetadata
	Spec         DocSchema
}

type DocMetadata struct {
	Name     string
	Group    string
	Scope    string
	Versions []string
}

type DocSchema struct {
	Description string
	Fields      []DocField
}

type DocField struct {
	Name        string
	Type        string
	Description string
	Required    bool
	Default     string
	Enum        []string
	Fields      []DocField // Nested fields
}

// Generate generates documentation for the given CRD in the specified format.
func (g *Generator) Generate(crd models.APICRD, format string) ([]byte, error) {
	data, err := g.Parse(crd)
	if err != nil {
		return nil, err
	}

	var tmplStr string
	switch format {
	case "markdown", "md":
		tmplStr = MarkdownTemplate
	case "html":
		tmplStr = HTMLTemplate
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	tmpl, err := template.New("doc").Parse(tmplStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}

// Parse extracts documentation data from the CRD.
func (g *Generator) Parse(crd models.APICRD) (DocData, error) {
	// Find the storage version or the first version to get the schema
	var schema *apiextensionsv1.JSONSchemaProps
	var versions []string

	for _, v := range crd.Spec.Versions {
		versions = append(versions, v.Name)
		if v.Storage {
			if v.Schema != nil && v.Schema.OpenAPIV3Schema != nil {
				schema = v.Schema.OpenAPIV3Schema
			}
		}
	}

	// Fallback if storage version doesn't have schema (unlikely but possible in some valid CRDs that use global schema in older versions, though v1 requires per-version)
	if schema == nil && len(crd.Spec.Versions) > 0 {
		if crd.Spec.Versions[0].Schema != nil && crd.Spec.Versions[0].Schema.OpenAPIV3Schema != nil {
			schema = crd.Spec.Versions[0].Schema.OpenAPIV3Schema
		}
	}

	if schema == nil {
		return DocData{}, fmt.Errorf("could not find OpenAPI V3 schema in CRD")
	}

	docSchema := g.parseSchema(*schema)

	return DocData{
		APIVersion:   crd.APIVersion,
		Kind:         crd.Kind,
		ResourceKind: crd.Spec.Names.Kind,
		Metadata: DocMetadata{
			Name:     crd.Metadata.Name,
			Group:    crd.Spec.Group,
			Scope:    string(crd.Spec.Scope),
			Versions: versions,
		},
		Spec: docSchema,
	}, nil
}

func (g *Generator) parseSchema(schema apiextensionsv1.JSONSchemaProps) DocSchema {
	return DocSchema{
		Description: schema.Description,
		Fields:      g.parseFields(schema.Properties, schema.Required),
	}
}

func (g *Generator) parseFields(properties map[string]apiextensionsv1.JSONSchemaProps, requiredFields []string) []DocField {
	var fields []DocField

	// Sort keys for deterministic output
	var keys []string
	for k := range properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		prop := properties[k]

		isRequired := false
		for _, req := range requiredFields {
			if req == k {
				isRequired = true
				break
			}
		}

		field := DocField{
			Name:        k,
			Type:        prop.Type,
			Description: prop.Description,
			Required:    isRequired,
		}

		if prop.Default != nil {
			field.Default = string(prop.Default.Raw) // Needs better formatting potentially
		}

		if len(prop.Enum) > 0 {
			for _, e := range prop.Enum {
				field.Enum = append(field.Enum, string(e.Raw))
			}
		}

		// Handle arrays
		if prop.Type == "array" && prop.Items != nil {
			if prop.Items.Schema != nil {
				field.Type = fmt.Sprintf("[]%s", prop.Items.Schema.Type)
				// If array of objects, parse nested fields
				if prop.Items.Schema.Type == "object" {
					field.Fields = g.parseFields(prop.Items.Schema.Properties, prop.Items.Schema.Required)
				}
			}
		} else if prop.Type == "object" {
			// Handle objects
			field.Fields = g.parseFields(prop.Properties, prop.Required)
			if prop.AdditionalProperties != nil && prop.AdditionalProperties.Schema != nil {
				field.Type = fmt.Sprintf("map[string]%s", prop.AdditionalProperties.Schema.Type)
			}
		}

		fields = append(fields, field)
	}

	return fields
}

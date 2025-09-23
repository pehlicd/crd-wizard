package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pehlicd/crd-wizard/internal/k8s"
	"github.com/pehlicd/crd-wizard/internal/logger"
	"gopkg.in/yaml.v2"
)

const (
	maxScannerCapacity = 1 * 1024 * 1024
)

type Config struct {
	OllamaHost     string
	OllamaModel    string
	RequestTimeout time.Time
}
type Client struct {
	OllamaURL  string
	Model      string
	HTTPClient *http.Client
	KubeClient *k8s.Client
	log        *logger.Logger
}

func NewClient(c Config, kubeClient *k8s.Client, l *logger.Logger) *Client {
	return &Client{
		OllamaURL: c.OllamaHost,
		Model:     c.OllamaModel,
		HTTPClient: &http.Client{
			Timeout: time.Until(c.RequestTimeout),
		},
		KubeClient: kubeClient,
		log:        l,
	}
}

// GenerateCrdContext performs the full RAG pipeline to generate documentation for a CRD.
func (c *Client) GenerateCrdContext(ctx context.Context, group, version, kind, schemaJSON string) (string, error) {
	// === 1. RETRIEVE (from Kubernetes) ===
	c.log.Info("Step 1: Retrieving live examples from the cluster...")
	crdExamples, err := c.KubeClient.FetchCRDExamples(ctx, group, version, kind)
	if err != nil {
		c.log.Warn("failed to fetch live examples from cluster", "err", err)
	}

	// === 2. AUGMENT ===
	c.log.Info("Step 2: Augmenting prompt with schema and examples...")
	prunedSchema, err := pruneSchema(schemaJSON)
	if err != nil {
		return "", fmt.Errorf("error pruning schema: %w", err)
	}
	prunedSchemaJSON, err := json.Marshal(prunedSchema)
	if err != nil {
		return "", fmt.Errorf("error marshaling pruned schema: %w", err)
	}

	// NEW: If no examples are found, generate a skeleton from the schema as a fallback.
	var skeletonYAML string
	if crdExamples == "" {
		c.log.Info("No live examples found; generating skeleton from schema as fallback.")
		skeletonYAML, err = generateYAMLFromSchema(group, version, kind, string(prunedSchemaJSON))
		if err != nil {
			// This is not a fatal error. The LLM can still try with just the schema.
			c.log.Warn("Failed to generate skeleton from schema", "err", err)
		}
	}

	prompt := c.buildAugmentedPrompt(group, version, kind, string(prunedSchemaJSON), crdExamples, skeletonYAML)

	// === 3. GENERATE ===
	c.log.Info("Step 3: Generating response from Ollama...")
	payload := map[string]any{
		"model":   c.Model,
		"prompt":  prompt,
		"system":  "You are an expert Kubernetes assistant specializing in CRD documentation. Your primary goal is to provide accurate, valid, and concise information based strictly on the context provided. You must never invent fields or values.",
		"stream":  true,
		"options": map[string]float64{"temperature": 0.1},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("error marshalling ollama payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", c.OllamaURL+"/api/generate", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("error creating ollama request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request to ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	c.log.Info("Processing streaming response from Ollama...")
	return c.processStreamingResponse(resp.Body)
}

// buildAugmentedPrompt now creates a context-sensitive prompt.
func (c *Client) buildAugmentedPrompt(group, version, kind, schemaJSON, examples, skeleton string) string {
	var contextSection string
	var instruction string
	// By default, we show the schema. This will be turned off if a skeleton is used.
	showSchema := true

	if examples != "" {
		// Strongest context: Use live examples
		instruction = "Base your example YAML primarily on the **Reference Examples from the Cluster** to ensure it is valid and realistic for that environment. The YAML must strictly adhere to the provided schema."
		contextSection = fmt.Sprintf(`
**Reference Examples from the Cluster (Primary Source):**
These are live examples. Use them as the definitive source for generating the YAML manifest's structure and realistic values.

%s
---`, examples)
	} else if skeleton != "" {
		// Fallback context: Use a generated skeleton.
		instruction = "Your ONLY task is to complete the YAML manifest below by replacing the placeholder values (e.g., 'value-for-some-field', 8080) with realistic and valid examples. DO NOT add, remove, or restructure any part of the YAML. The structure provided is rigid and MUST be followed exactly."
		contextSection = fmt.Sprintf(`
**YAML Skeleton to Complete:**
This skeleton was generated from the CRD schema because no live examples were found.
You must complete it without altering its structure.

%s
---`, skeleton)
		showSchema = false
	} else {
		// Weakest context: Schema only
		instruction = "Generate a valid example YAML manifest based ONLY on the provided schema. Be conservative and only include required fields and a few important optional ones."
		contextSection = ""
	}

	schemaDetailsSection := ""
	if showSchema {
		schemaDetailsSection = fmt.Sprintf(`
**Input Schema Details:**
- Group: %s
- Version: %s
- Kind: %s
- Pruned OpenAPI v3 Schema (JSON):
%s`, group, version, kind, schemaJSON)
	}

	return fmt.Sprintf(`
You are an expert Kubernetes assistant. Your task is to generate a concise explanation and a valid example YAML manifest.

**INSTRUCTIONS:**
%s

**CONTEXT:**
%s%s

---
**OUTPUT:**

**1. Explanation:**
(A brief summary, 2-4 sentences, of the resource's purpose)

**2. Example YAML Manifest:**
(A single, complete YAML manifest in a markdown code block. It must be a NEW example, not a copy.)
- 'apiVersion' MUST be '%s/%s'.
- 'kind' MUST be '%s'.
- 'metadata.name' MUST be a new, realistic name (e.g., "my-%s-sample").
`, instruction, contextSection, schemaDetailsSection, group, version, kind, strings.ToLower(kind))
}

// generateYAMLFromSchema creates a YAML manifest skeleton from a JSON schema.
func generateYAMLFromSchema(group, version, kind, schemaJSON string) (string, error) {
	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
		return "", fmt.Errorf("failed to unmarshal schema for skeleton generation: %w", err)
	}

	root := make(map[string]interface{})
	root["apiVersion"] = fmt.Sprintf("%s/%s", group, version)
	root["kind"] = kind
	root["metadata"] = map[string]interface{}{"name": fmt.Sprintf("my-%s-sample", strings.ToLower(kind))}

	if props, ok := schema["properties"].(map[string]interface{}); ok {
		if specSchema, ok := props["spec"].(map[string]interface{}); ok {
			root["spec"] = buildObjectFromSchema(specSchema)
		}
	}

	yamlBytes, err := yaml.Marshal(root)
	if err != nil {
		return "", fmt.Errorf("failed to marshal skeleton to YAML: %w", err)
	}

	return string(yamlBytes), nil
}

// buildObjectFromSchema recursively builds a map based on schema properties.
func buildObjectFromSchema(schema map[string]interface{}) map[string]interface{} {
	obj := make(map[string]interface{})
	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		return obj
	}

	requiredSet := make(map[string]bool)
	if required, ok := schema["required"].([]interface{}); ok {
		for _, req := range required {
			if r, ok := req.(string); ok {
				requiredSet[r] = true
			}
		}
	}

	// We only include required fields in the skeleton to keep it minimal and accurate.
	for key, val := range properties {
		if requiredSet[key] {
			if propSchema, ok := val.(map[string]interface{}); ok {
				obj[key] = generateValueForSchema(key, propSchema)
			}
		}
	}
	return obj
}

// generateValueForSchema creates a more intelligent placeholder value for a given schema type.
func generateValueForSchema(key string, propSchema map[string]interface{}) interface{} {
	propType, _ := propSchema["type"].(string)
	lowerKey := strings.ToLower(key)

	switch propType {
	case "string":
		if strings.Contains(lowerKey, "image") {
			return "repository/image-name:tag"
		}
		if strings.Contains(lowerKey, "host") {
			return "service.namespace.svc.cluster.local"
		}
		if strings.Contains(lowerKey, "version") {
			return "v1.0.0"
		}
		if strings.Contains(lowerKey, "name") && lowerKey != "name" { // metadata.name is handled separately
			return fmt.Sprintf("my-%s", key)
		}
		if strings.Contains(lowerKey, "storageclass") {
			return "standard"
		}
		return fmt.Sprintf("value-for-%s", key)

	case "integer", "number":
		if strings.Contains(lowerKey, "port") {
			return 8080
		}
		if strings.Contains(lowerKey, "replica") {
			return 3
		}
		return 1

	case "boolean":
		return true

	case "object":
		return buildObjectFromSchema(propSchema)

	case "array":
		var itemsSchema map[string]interface{}
		if items, ok := propSchema["items"].(map[string]interface{}); ok {
			itemsSchema = items
		}
		if itemsSchema == nil {
			return []interface{}{}
		}
		// Pass a generic key for items inside an array
		return []interface{}{generateValueForSchema("item", itemsSchema)}

	default:
		return fmt.Sprintf("placeholder-for-%s", key)
	}
}

// --- Existing Helper Functions (Unchanged) ---

// processStreamingResponse reads the streaming response from Ollama and concatenates it.
func (c *Client) processStreamingResponse(body io.Reader) (string, error) {
	var fullResponse strings.Builder
	scanner := bufio.NewScanner(body)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, maxScannerCapacity)

	for scanner.Scan() {
		line := scanner.Text()
		var streamResp struct {
			Response string `json:"response"`
			Done     bool   `json:"done"`
		}
		if err := json.Unmarshal([]byte(line), &streamResp); err != nil {
			c.log.Warn("failed to unmarshal ollama stream line", "line", line, "err", err)
			continue
		}
		fullResponse.WriteString(streamResp.Response)
		if streamResp.Done {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading ollama response stream: %w", err)
	}

	return fullResponse.String(), nil
}

// pruneSchema recursively removes all fields from the schema except for a whitelist.
func pruneSchema(schemaJSON string) (map[string]any, error) {
	var schema map[string]any
	if err := json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema for pruning: %w", err)
	}
	return pruneMap(schema), nil
}

// pruneMap is the recursive helper for pruneSchema.
func pruneMap(data map[string]any) map[string]any {
	if data == nil {
		return nil
	}
	whitelist := map[string]bool{
		"properties":  true,
		"type":        true,
		"items":       true,
		"description": true,
		"required":    true,
	}

	result := make(map[string]any)
	for key, val := range data {
		if !whitelist[key] {
			continue
		}
		switch v := val.(type) {
		case map[string]any:
			result[key] = pruneMap(v)
		case []any:
			var newArr []any
			for _, item := range v {
				if itemMap, ok := item.(map[string]any); ok {
					newArr = append(newArr, pruneMap(itemMap))
				} else {
					newArr = append(newArr, item)
				}
			}
			result[key] = newArr
		default:
			result[key] = v
		}
	}
	return result
}

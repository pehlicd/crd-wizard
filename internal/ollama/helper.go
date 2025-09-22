package ollama

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
)

const (
	// defaultOllamaHost is the default address for the Ollama API.
	defaultOllamaHost = "http://localhost:11435"
	// defaultOllamaModel is the default model to use for generation.
	defaultOllamaModel = "tinyllama"
	// defaultRequestTimeout is the default timeout for HTTP requests to the Ollama API.
	defaultRequestTimeout = 10 * time.Minute
	maxScannerCapacity    = 1 * 1024 * 1024
)

// Client holds the configuration for the Ollama client
type Client struct {
	OllamaURL  string
	Model      string
	HTTPClient *http.Client
}

// NewClient creates a new Ollama client with default settings
func NewClient() *Client {
	return &Client{
		OllamaURL: defaultOllamaHost,
		Model:     defaultOllamaModel,
		HTTPClient: &http.Client{
			Timeout: defaultRequestTimeout,
		},
	}
}

// GenerateCrdContext sends a CRD schema to Ollama and gets a response
func (c *Client) GenerateCrdContext(ctx context.Context, group, version, kind, schemaJSON string) (string, error) {
	// First, prune the schema to reduce its size
	prunedSchema, err := pruneSchema(schemaJSON)
	if err != nil {
		return "", fmt.Errorf("error pruning schema: %w", err)
	}
	prunedSchemaJSON, err := json.Marshal(prunedSchema)
	if err != nil {
		return "", fmt.Errorf("error marshaling pruned schema: %w", err)
	}

	prompt := c.buildPrompt(group, version, kind, string(prunedSchemaJSON))

	payload := map[string]any{
		"model":  c.Model,
		"prompt": prompt,
		"stream": true, // Using streaming
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

	return c.processStreamingResponse(resp.Body)
}

func (c *Client) buildPrompt(group, version, kind, schemaJSON string) string {
	// Enhanced prompt with stricter instructions and a clearer structure.
	return fmt.Sprintf(`
Analyze the provided Kubernetes CRD schema. Your task is to generate a concise explanation and a valid example YAML manifest.

**Input Schema Details:**
- Group: %s
- Version: %s
- Kind: %s
- Pruned OpenAPI v3 Schema (JSON):
%s

---

**Output Requirements:**

**1. Explanation:**
- Provide a brief summary (2-4 sentences) of this Custom Resource's purpose.
- Explain what problem it solves or what it represents in a cluster.

**2. Example YAML Manifest:**
- Generate one single, complete, and valid YAML manifest for an example resource.
- The YAML MUST strictly adhere to the provided schema.
- It MUST include 'apiVersion', 'kind', 'metadata', and a 'spec' section.
- 'apiVersion' MUST be exactly '%s/%s'.
- 'kind' MUST be exactly '%s'.
- 'metadata.name' MUST be a realistic, lowercase example name (e.g., "my-%s-sample").
- All fields in the 'spec' should have realistic, illustrative example values, not generic placeholders like "your-value-here" or "string".
- The entire YAML manifest MUST be enclosed in a single markdown code block.

**IMPORTANT RULES:**
- Do not repeat any of these instructions or the input schema in your response.
- Provide the explanation text first, followed immediately by the YAML code block.
- There MUST be NO text, comments, or explanations after the final YAML code block.
`, group, version, kind, schemaJSON, group, version, kind, strings.ToLower(kind))
}

// processStreamingResponse reads the streaming response from Ollama and concatenates it.
func (c *Client) processStreamingResponse(body io.Reader) (string, error) {
	var fullResponse strings.Builder
	scanner := bufio.NewScanner(body)

	// Create a buffer and set the scanner to use it.
	buf := make([]byte, 0, 64*1024) // 64KB initial buffer
	scanner.Buffer(buf, maxScannerCapacity)

	for scanner.Scan() {
		line := scanner.Text()
		var streamResp struct {
			Response string `json:"response"`
			Done     bool   `json:"done"`
		}
		if err := json.Unmarshal([]byte(line), &streamResp); err != nil {
			return "", fmt.Errorf("error decoding ollama stream line: %w", err)
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
func pruneSchema(schemaJSON string) (map[string]interface{}, error) {
	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema for pruning: %w", err)
	}

	pruned := pruneMap(schema)
	return pruned, nil
}

// pruneMap is the recursive helper for pruneSchema.
func pruneMap(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		return nil
	}
	// Whitelist of keys to keep at each level of the schema.
	whitelist := map[string]bool{
		"properties":  true,
		"type":        true,
		"items":       true,
		"description": true,
		"required":    true,
	}

	result := make(map[string]interface{})
	for key, val := range data {
		if !whitelist[key] {
			continue // Skip non-whitelisted keys
		}

		switch v := val.(type) {
		case map[string]interface{}:
			result[key] = pruneMap(v)
		case []interface{}: // This could be 'required' array or other arrays
			// Check if it's an array of objects to recurse into
			var newArr []interface{}
			for _, item := range v {
				if itemMap, ok := item.(map[string]interface{}); ok {
					newArr = append(newArr, pruneMap(itemMap))
				} else {
					newArr = append(newArr, item) // Keep primitive values (like in 'required' array)
				}
			}
			result[key] = newArr
		default:
			result[key] = v
		}
	}
	return result
}

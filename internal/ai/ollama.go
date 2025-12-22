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
)

const (
	maxScannerCapacity = 2 * 1024 * 1024 // Increased buffer for larger responses
)

type OllamaProvider struct {
	Config     Config
	HTTPClient *http.Client
}

func NewOllamaProvider(c Config, client *http.Client) *OllamaProvider {
	return &OllamaProvider{
		Config:     c,
		HTTPClient: client,
	}
}

func (p *OllamaProvider) Name() string {
	return string(AIProviderOllama)
}

// Generate handles the raw HTTP interaction with Ollama
func (p *OllamaProvider) Generate(ctx context.Context, prompt string) (string, error) {
	options := map[string]any{
		"temperature": 0.2,
		"top_p":       0.9,
	}
	if p.Config.OllamaNumCtx > 0 {
		options["num_ctx"] = p.Config.OllamaNumCtx
	}

	payload := map[string]any{
		"model":   p.Config.Model,
		"prompt":  prompt,
		"system":  "You are a Senior Kubernetes Engineer. Your output must be technical, precise, and valid YAML. Do not chat. Do not provide preamble like 'Here is the file'. Output Markdown only.",
		"stream":  true,
		"options": options,
	}

	if p.Config.OllamaKeepAlive != "" {
		payload["keep_alive"] = p.Config.OllamaKeepAlive
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("error marshalling payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.Config.OllamaHost+"/api/generate", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request to ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama request failed (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	return p.processStreamingResponse(resp.Body)
}

func (p *OllamaProvider) processStreamingResponse(body io.Reader) (string, error) {
	var fullResponse strings.Builder
	fullResponse.Grow(4096)

	scanner := bufio.NewScanner(body)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, maxScannerCapacity)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var streamResp struct {
			Response string `json:"response"`
			Done     bool   `json:"done"`
		}
		if err := json.Unmarshal(line, &streamResp); err != nil {
			continue
		}
		fullResponse.WriteString(streamResp.Response)
		if streamResp.Done {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading stream: %w", err)
	}
	return fullResponse.String(), nil
}

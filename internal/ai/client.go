package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v2"

	"github.com/pehlicd/crd-wizard/internal/k8s"
	"github.com/pehlicd/crd-wizard/internal/logger"
)

const (
	googleSearchAPI = "https://www.googleapis.com/customsearch/v1"
	ddgSearchURL    = "https://html.duckduckgo.com/html/"
)

type Client struct {
	Config     Config
	HTTPClient *http.Client
	KubeClient *k8s.Client
	log        *logger.Logger
	Provider   LLMProvider

	// Cache storage
	cacheMu sync.RWMutex
	cache   map[string]string
}

func NewClient(c Config, kubeClient *k8s.Client, l *logger.Logger) *Client {
	if c.RequestTimeout == 0 {
		c.RequestTimeout = 2 * time.Minute
	}
	// Default to Ollama if no provider specified
	if c.Provider == "" {
		c.Provider = ProviderOllama
	}
	// Default to DDG if search is enabled but no provider specified
	if c.EnableSearch && c.SearchProvider == "" {
		c.SearchProvider = SearchProviderDuckDuckGo
	}
	// Default retries
	if c.MaxValidationRetries == 0 {
		c.MaxValidationRetries = 10
	}

	httpClient := &http.Client{
		Timeout: c.RequestTimeout,
	}

	var provider LLMProvider
	switch c.Provider {
	case ProviderOllama:
		provider = NewOllamaProvider(c, httpClient)
	case ProviderGemini:
		var err error
		provider, err = NewGeminiProvider(context.Background(), c.GeminiAPIKey, c.Model)
		if err != nil {
			l.Warn("failed to initialize gemini provider, falling back to ollama", "err", err)
			provider = NewOllamaProvider(c, httpClient)
		}
	default:
		// Fallback to Ollama or error? For now, fallback or panic if strict
		l.Warn("Unknown provider, falling back to Ollama", "provider", c.Provider)
		provider = NewOllamaProvider(c, httpClient)
	}

	client := &Client{
		Config:     c,
		HTTPClient: httpClient,
		KubeClient: kubeClient,
		log:        l,
		Provider:   provider,
	}

	// Initialize cache if enabled
	if c.EnableCache {
		client.cache = make(map[string]string)
	}

	return client
}

// GenerateCrdContext performs the full RAG pipeline to generate documentation for a CRD.
func (c *Client) GenerateCrdContext(ctx context.Context, group, version, kind, schemaJSON string) (string, error) {
	// 1. Check Cache (Fast Path)
	cacheKey := fmt.Sprintf("%s/%s/%s", group, version, kind)
	if c.Config.EnableCache {
		c.cacheMu.RLock()
		val, found := c.cache[cacheKey]
		c.cacheMu.RUnlock()
		if found {
			c.log.Info("Serving CRD documentation from cache", "key", cacheKey)
			return val, nil
		}
	}

	g, groupCtx := errgroup.WithContext(ctx)

	var (
		crdExamples string
		webResults  string
	)

	// Task A: Fetch Live Examples from K8s
	g.Go(func() error {
		c.log.Info("retrieving live examples from cluster")
		ex, err := c.KubeClient.FetchCRDExamples(groupCtx, group, version, kind)
		if err != nil {
			c.log.Warn("failed to fetch live examples", "err", err)
			return nil // Non-fatal
		}
		crdExamples = ex
		return nil
	})

	// Task B: Perform Web Search (If enabled)
	if c.Config.EnableSearch {
		g.Go(func() error {
			c.log.Info(fmt.Sprintf("searching web using %s", c.Config.SearchProvider))
			query := fmt.Sprintf("kubernetes crd %s %s %s example yaml", group, version, kind)

			var res string
			var err error

			if c.Config.SearchProvider == SearchProviderGoogle {
				res, err = c.performGoogleSearch(groupCtx, query)
			} else {
				res, err = c.performDuckDuckGoSearch(groupCtx, query)
			}

			if err != nil {
				c.log.Warn("web search failed", "provider", c.Config.SearchProvider, "err", err)
				return nil // Non-fatal
			}
			webResults = res
			return nil
		})
	}

	// Task C: Prune Schema (CPU bound, run locally)
	c.log.Info("pruning schema")
	prunedSchema, err := pruneSchema(schemaJSON)
	if err != nil {
		return "", fmt.Errorf("error pruning schema: %w", err)
	}
	prunedSchemaJSON, err := json.Marshal(prunedSchema)
	if err != nil {
		return "", fmt.Errorf("error marshaling pruned schema: %w", err)
	}

	// Wait for network tasks to finish
	if err := g.Wait(); err != nil {
		return "", err
	}

	// Logic: Fallback generation if no live examples found
	var skeletonYAML string
	if crdExamples == "" {
		c.log.Info("No live examples found; generating skeleton from schema.")
		skeletonYAML, err = generateYAMLFromSchema(group, version, kind, string(prunedSchemaJSON))
		if err != nil {
			c.log.Warn("Failed to generate skeleton", "err", err)
		}
	}

	basePrompt := c.buildAugmentedPrompt(group, version, kind, string(prunedSchemaJSON), crdExamples, skeletonYAML, webResults)
	currentPrompt := basePrompt
	var finalResponse string

	for attempt := 0; attempt <= c.Config.MaxValidationRetries; attempt++ {
		c.log.Info("generating response from AI provider", "provider", c.Provider.Name(), "attempt", attempt+1)

		response, err := c.Provider.Generate(ctx, currentPrompt)
		if err != nil {
			return "", err
		}

		// Validation Step
		c.log.Info("validating generated example via dry-run")
		validationErr := c.validateGeneratedContent(ctx, response)
		if validationErr == nil {
			c.log.Info("validation successful")
			finalResponse = response
			break
		}

		c.log.Warn("validation failed", "err", validationErr)

		// If this was the last attempt, return the best we have (or error out)
		if attempt == c.Config.MaxValidationRetries {
			c.log.Warn("max retries reached, returning last response despite validation errors")
			finalResponse = response
			// Optional: append a warning to the final response
			finalResponse += fmt.Sprintf("\n\n> **Warning:** Automatic validation failed: %v", validationErr)
			break
		}

		// Update prompt for next iteration with the error
		currentPrompt = c.buildCorrectionPrompt(basePrompt, response, validationErr.Error())
	}

	// Save to Cache
	if c.Config.EnableCache && finalResponse != "" {
		c.cacheMu.Lock()
		c.cache[cacheKey] = finalResponse
		c.cacheMu.Unlock()
	}

	return finalResponse, nil
}

// validateGeneratedContent extracts YAML and calls the K8s dry-run
func (c *Client) validateGeneratedContent(ctx context.Context, content string) error {
	yamlContent := extractYAMLBlock(content)
	if yamlContent == "" {
		return fmt.Errorf("no yaml block found in response")
	}

	return c.KubeClient.DryRun(ctx, yamlContent)
}

func extractYAMLBlock(content string) string {
	// Look for standard markdown code blocks
	re := regexp.MustCompile("(?s)```yaml\\s+(.*?)\\s+```")
	match := re.FindStringSubmatch(content)
	if len(match) > 1 {
		return match[1]
	}

	// Fallback: Look for generic code blocks if yaml tag is missing
	reGeneric := regexp.MustCompile("(?s)```\\s+(.*?)\\s+```")
	matchGeneric := reGeneric.FindStringSubmatch(content)
	if len(matchGeneric) > 1 {
		return matchGeneric[1]
	}

	// Fallback: If the model returned pure YAML without markdown
	if strings.Contains(content, "apiVersion:") && strings.Contains(content, "kind:") {
		return content
	}

	return ""
}

func (c *Client) buildCorrectionPrompt(originalPrompt, previousResponse, errorMsg string) string {
	var sb strings.Builder
	// We reset to the original request but append the failure context
	sb.WriteString(originalPrompt)
	sb.WriteString("\n\n")
	sb.WriteString("!!! CRITICAL: VALIDATION FAILED !!!\n")
	sb.WriteString("The YAML you generated in the previous attempt was invalid:\n")
	sb.WriteString("<invalid_generation>\n")
	sb.WriteString(previousResponse)
	sb.WriteString("\n</invalid_generation>\n\n")
	sb.WriteString("Kubernetes API Server Error:\n")
	sb.WriteString(fmt.Sprintf("`%s`\n\n", errorMsg))
	sb.WriteString("Please regenerate the ENTIRE response. Fix the YAML to satisfy the schema and validation error above.")
	return sb.String()
}

// performDuckDuckGoSearch scrapes the HTML version of DuckDuckGo (No API Key needed)
func (c *Client) performDuckDuckGoSearch(ctx context.Context, query string) (string, error) {
	data := url.Values{}
	data.Set("q", query)
	data.Set("kl", "us-en")

	req, err := http.NewRequestWithContext(ctx, "POST", ddgSearchURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ddg returned status %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	htmlContent := string(bodyBytes)

	reLink := regexp.MustCompile(`<a[^>]+class="result__a"[^>]+href="([^"]+)"[^>]*>(.*?)</a>`)
	reSnippet := regexp.MustCompile(`<a[^>]+class="result__snippet"[^>]*>(.*?)</a>`)

	links := reLink.FindAllStringSubmatch(htmlContent, 5)
	snippets := reSnippet.FindAllStringSubmatch(htmlContent, 5)

	if len(links) == 0 {
		return "", fmt.Errorf("no results found on ddg")
	}

	var sb strings.Builder
	sb.WriteString("Source: DuckDuckGo (Web)\n")

	count := 0
	for i, match := range links {
		if count >= 3 {
			break
		}
		if len(match) < 3 {
			continue
		}

		urlVal := match[1]
		title := stripTags(match[2])
		snippetVal := ""
		if i < len(snippets) && len(snippets[i]) >= 2 {
			snippetVal = stripTags(snippets[i][1])
		}

		if decoded, err := url.QueryUnescape(urlVal); err == nil {
			urlVal = decoded
		}

		sb.WriteString(fmt.Sprintf("- Title: %s\n  Link: %s\n  Snippet: %s\n", title, urlVal, snippetVal))
		count++
	}

	return sb.String(), nil
}

func stripTags(content string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(content, "")
}

func (c *Client) performGoogleSearch(ctx context.Context, query string) (string, error) {
	if c.Config.GoogleAPIKey == "" || c.Config.GoogleCX == "" {
		return "", fmt.Errorf("google search enabled but credentials missing")
	}

	u, _ := url.Parse(googleSearchAPI)
	q := u.Query()
	q.Set("key", c.Config.GoogleAPIKey)
	q.Set("cx", c.Config.GoogleCX)
	q.Set("q", query)
	q.Set("num", "3")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return "", err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("search api returned %d", resp.StatusCode)
	}

	var searchResp struct {
		Items []struct {
			Title   string `json:"title"`
			Snippet string `json:"snippet"`
			Link    string `json:"link"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("Source: Google API\n")
	for _, item := range searchResp.Items {
		sb.WriteString(fmt.Sprintf("- Title: %s\n  Link: %s\n  Snippet: %s\n", item.Title, item.Link, item.Snippet))
	}

	return sb.String(), nil
}

func (c *Client) buildAugmentedPrompt(group, version, kind, schemaJSON, examples, skeleton, webResults string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Create a production-ready Kubernetes YAML manifest for Kind: `%s` (Group: `%s`, Version: `%s`).\n\n", kind, group, version))

	sb.WriteString("Use the following context to ensure accuracy:\n")

	if webResults != "" {
		sb.WriteString("<web_search_results>\n")
		sb.WriteString("Use these snippets to understand the purpose of fields:\n")
		sb.WriteString(webResults)
		sb.WriteString("\n</web_search_results>\n\n")
	}

	if examples != "" {
		sb.WriteString("<live_cluster_examples>\n")
		sb.WriteString("HIGH PRIORITY. Mimic the structure and values found here:\n")
		sb.WriteString(examples)
		sb.WriteString("\n</live_cluster_examples>\n\n")
	} else if skeleton != "" {
		sb.WriteString("<schema_skeleton>\n")
		sb.WriteString("No live examples found. Fill in this skeleton with realistic values:\n")
		sb.WriteString(skeleton)
		sb.WriteString("\n</schema_skeleton>\n\n")
	}

	sb.WriteString("<openapi_schema>\n")
	sb.WriteString(schemaJSON)
	sb.WriteString("\n</openapi_schema>\n\n")

	sb.WriteString(`
**COMMANDS:**
1. **Analyze**: Briefly explain the resource's purpose based on the search results and schema (max 3 sentences).
2. **Generate**: Provide ONE complete YAML manifest.
   - Use 'apiVersion: ` + fmt.Sprintf("%s/%s", group, version) + `'
   - Use 'kind: ` + kind + `'
   - Do NOT use placeholders like 'string' or 'value'. Use realistic defaults (e.g., port: 80, image: nginx:latest).
   - If <live_cluster_examples> are provided, prefer their configuration style.

**OUTPUT FORMAT:**
### Explanation
(Text here)

### Manifest
` + "```yaml" + `
(YAML here)
` + "```" + `
`)

	return sb.String()
}

func generateYAMLFromSchema(group, version, kind, schemaJSON string) (string, error) {
	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
		return "", fmt.Errorf("failed to unmarshal schema: %w", err)
	}

	root := make(map[string]interface{})
	root["apiVersion"] = fmt.Sprintf("%s/%s", group, version)
	root["kind"] = kind
	root["metadata"] = map[string]interface{}{"name": fmt.Sprintf("my-%s-demo", strings.ToLower(kind))}

	if props, ok := schema["properties"].(map[string]interface{}); ok {
		if specSchema, ok := props["spec"].(map[string]interface{}); ok {
			root["spec"] = buildObjectFromSchema(specSchema)
		}
	}
	yamlBytes, err := yaml.Marshal(root)
	if err != nil {
		return "", fmt.Errorf("failed to marshal skeleton: %w", err)
	}
	return string(yamlBytes), nil
}

func buildObjectFromSchema(schema map[string]interface{}) map[string]interface{} {
	obj := make(map[string]interface{})
	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		return obj
	}
	requiredSet := make(map[string]struct{})
	if required, ok := schema["required"].([]interface{}); ok {
		for _, req := range required {
			if r, ok := req.(string); ok {
				requiredSet[r] = struct{}{}
			}
		}
	}
	for key, val := range properties {
		if _, isRequired := requiredSet[key]; isRequired {
			if propSchema, ok := val.(map[string]interface{}); ok {
				obj[key] = generateValueForSchema(key, propSchema)
			}
		}
	}
	return obj
}

func generateValueForSchema(key string, propSchema map[string]interface{}) interface{} {
	propType, _ := propSchema["type"].(string)
	lowerKey := strings.ToLower(key)

	switch propType {
	case "string":
		if strings.Contains(lowerKey, "image") {
			return "nginx:latest"
		}
		if strings.Contains(lowerKey, "host") {
			return "example.com"
		}
		return fmt.Sprintf("example-%s", key)
	case "integer", "number":
		if strings.Contains(lowerKey, "port") {
			return 8080
		}
		if strings.Contains(lowerKey, "replica") {
			return 2
		}
		return 1
	case "boolean":
		return false
	case "object":
		return buildObjectFromSchema(propSchema)
	case "array":
		if items, ok := propSchema["items"].(map[string]interface{}); ok {
			return []interface{}{generateValueForSchema("item", items)}
		}
		return []interface{}{}
	default:
		return "value"
	}
}

func pruneSchema(schemaJSON string) (map[string]any, error) {
	var schema map[string]any
	if err := json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
	}
	return pruneMap(schema, 0), nil
}

// pruneMap aggressively trims the schema to fit within LLM context limits.
func pruneMap(data map[string]any, depth int) map[string]any {
	if data == nil {
		return nil
	}
	// Hard depth limit. K8s CRDs can be deeply nested.
	// Depth 10 is usually enough for the core structure.
	if depth > 10 {
		return map[string]any{"type": "object", "description": "truncated-depth"}
	}

	result := make(map[string]any)

	// Whitelist of keys to keep
	keysToKeep := []string{"type", "required", "items", "properties", "x-kubernetes-int-or-string"}

	// Only keep descriptions at top level or very shallow levels, and truncate them heavily
	if depth < 3 {
		if desc, ok := data["description"].(string); ok {
			if len(desc) > 100 {
				result["description"] = desc[:97] + "..."
			} else {
				result["description"] = desc
			}
		}
	}

	for _, k := range keysToKeep {
		val, exists := data[k]
		if !exists {
			continue
		}

		switch k {
		case "properties":
			if props, ok := val.(map[string]any); ok {
				newProps := make(map[string]any)
				for propName, propVal := range props {
					if propMap, ok := propVal.(map[string]any); ok {
						newProps[propName] = pruneMap(propMap, depth+1)
					}
				}
				if len(newProps) > 0 {
					result[k] = newProps
				}
			}
		case "items":
			if itemsMap, ok := val.(map[string]any); ok {
				result[k] = pruneMap(itemsMap, depth+1)
			}
		case "required", "type", "x-kubernetes-int-or-string":
			result[k] = val
		}
	}

	return result
}

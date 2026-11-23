package llm

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"golang.org/x/sync/semaphore"
)

const cacheSize = 128

// ─── Ollama REST API types ────────────────────────────────────────────────────

type ollamaGenerateRequest struct {
	Model   string         `json:"model"`
	Prompt  string         `json:"prompt"`
	Stream  bool           `json:"stream"`
	Options map[string]any `json:"options,omitempty"`
}

type ollamaGenerateChunk struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
	Error    string `json:"error,omitempty"`
}

type ollamaTagsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

// ─── Cloud / OpenAI-Compatible API types ───────────────────────────────────────

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Stream      bool          `json:"stream"`
	Temperature float64       `json:"temperature"`
}

type chatCompletionChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// ─── Unified Client ───────────────────────────────────────────────────────────

// Client wraps LLM streaming with concurrency control, multi-provider support, and LRU caching.
type Client struct {
	provider  string // "gemini", "groq", "openai", "cloud", "ollama"
	baseURL   string
	model     string
	apiKey    string
	http      *http.Client
	sem       *semaphore.Weighted
	cache     *lru.Cache[string, string]
	active    atomic.Int64
	maxActive int64
}

// Config specifies the client configuration.
type Config struct {
	Provider      string
	BaseURL       string
	Model         string
	APIKey        string
	MaxConcurrent int64
}

// NewClient initializes the LLM client with automatic provider detection.
func NewClient(cfg Config) (*Client, error) {
	cache, err := lru.New[string, string](cacheSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create LRU cache: %w", err)
	}

	httpClient := &http.Client{
		Timeout: 0, // No timeout — streaming LLM responses
	}

	if cfg.MaxConcurrent <= 0 {
		cfg.MaxConcurrent = 5
	}

	// Auto-detect provider if not explicitly given
	provider := strings.ToLower(strings.TrimSpace(cfg.Provider))
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	model := strings.TrimSpace(cfg.Model)
	apiKey := strings.TrimSpace(cfg.APIKey)

	if provider == "" || provider == "auto" {
		if k := os.Getenv("GEMINI_API_KEY"); k != "" {
			provider = "gemini"
			apiKey = k
		} else if k := os.Getenv("GROQ_API_KEY"); k != "" {
			provider = "groq"
			apiKey = k
		} else if k := os.Getenv("OPENAI_API_KEY"); k != "" {
			provider = "openai"
			apiKey = k
		} else if k := os.Getenv("LLM_API_KEY"); k != "" {
			provider = "cloud"
			apiKey = k
		} else {
			provider = "ollama"
		}
	}

	// Set defaults according to provider
	switch provider {
	case "gemini":
		if baseURL == "" {
			baseURL = "https://generativelanguage.googleapis.com/v1beta/openai"
		}
		if model == "" {
			model = getEnv("GEMINI_MODEL", "gemini-2.5-flash")
		}
		if apiKey == "" {
			apiKey = os.Getenv("GEMINI_API_KEY")
		}

	case "groq":
		if baseURL == "" {
			baseURL = "https://api.groq.com/openai/v1"
		}
		if model == "" {
			model = getEnv("GROQ_MODEL", "llama-3.3-70b-versatile")
		}
		if apiKey == "" {
			apiKey = os.Getenv("GROQ_API_KEY")
		}

	case "openai":
		if baseURL == "" {
			baseURL = "https://api.openai.com/v1"
		}
		if model == "" {
			model = getEnv("OPENAI_MODEL", "gpt-4o-mini")
		}
		if apiKey == "" {
			apiKey = os.Getenv("OPENAI_API_KEY")
		}

	case "cloud":
		if baseURL == "" {
			baseURL = getEnv("LLM_BASE_URL", "https://api.openai.com/v1")
		}
		if model == "" {
			model = getEnv("LLM_MODEL", "gpt-4o-mini")
		}
		if apiKey == "" {
			apiKey = os.Getenv("LLM_API_KEY")
		}

	default: // "ollama"
		provider = "ollama"
		if baseURL == "" {
			baseURL = getEnv("OLLAMA_URL", "http://127.0.0.1:11434")
		}
		if model == "" {
			model = getEnv("OLLAMA_MODEL", "qwen2.5-coder:7b")
		}
		model = resolveOllamaModel(baseURL, model)
	}

	return &Client{
		provider:  provider,
		baseURL:   baseURL,
		model:     model,
		apiKey:    apiKey,
		sem:       semaphore.NewWeighted(cfg.MaxConcurrent),
		cache:     cache,
		maxActive: cfg.MaxConcurrent,
		http:      httpClient,
	}, nil
}

// Provider returns the active LLM provider name.
func (c *Client) Provider() string {
	return c.provider
}

// Model returns the active model name.
func (c *Client) Model() string {
	return c.model
}

// IsCloud returns true if using a cloud API.
func (c *Client) IsCloud() bool {
	return c.provider != "ollama"
}

// Status returns readiness and a diagnostic message.
func (c *Client) Status() (bool, string) {
	if c.IsCloud() {
		if c.apiKey == "" {
			return false, fmt.Sprintf("%s API key missing. Set %s_API_KEY in backend/.env", strings.ToUpper(c.provider), strings.ToUpper(c.provider))
		}
		return true, fmt.Sprintf("Connected to %s (%s)", strings.ToUpper(c.provider), c.model)
	}

	// Ollama check
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/tags", nil)
	if err != nil {
		return false, "Ollama connection error"
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return false, "Ollama is not running. Start Ollama or set GEMINI_API_KEY / GROQ_API_KEY in backend/.env"
	}
	defer resp.Body.Close()
	return true, fmt.Sprintf("Connected to local Ollama (%s)", c.model)
}

// IsAtCapacity returns true if maximum concurrent requests are in-flight.
func (c *Client) IsAtCapacity() bool {
	return c.active.Load() >= c.maxActive
}

// WarmUp sends a warm-up request if using local Ollama.
func (c *Client) WarmUp() {
	if c.IsCloud() {
		return // Cloud APIs don't need local warm-up
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Printf("🔥 Checking Ollama model %s...\n", c.model)
	_ = c.streamGenerateOllama(ctx, "ping", map[string]any{"num_predict": 1}, func(_ string) error {
		return nil
	})
}

// StreamSimulation streams simulation code generation.
func (c *Client) StreamSimulation(
	ctx context.Context,
	userPrompt string,
	params map[string]any,
	onChunk func(token string) error,
) (string, error) {
	systemPrompt := BuildSystemPrompt()
	userMessage := BuildUserMessage(userPrompt, params)
	fullPrompt := systemPrompt + "\n\n" + userMessage

	// Cache hit — replay cached code instantly
	key := cacheKey(fullPrompt)
	if cached, ok := c.cache.Get(key); ok {
		return cached, onChunk(cached)
	}

	// Acquire semaphore slot
	if err := c.sem.Acquire(ctx, 1); err != nil {
		return "", fmt.Errorf("request cancelled while waiting for LLM slot: %w", err)
	}
	defer c.sem.Release(1)

	c.active.Add(1)
	defer c.active.Add(-1)

	var accumulated strings.Builder
	collectChunk := func(token string) error {
		accumulated.WriteString(token)
		return onChunk(token)
	}

	var err error
	if c.IsCloud() {
		err = c.streamGenerateCloud(ctx, systemPrompt, userMessage, collectChunk)
	} else {
		err = c.streamGenerateOllama(ctx, fullPrompt, map[string]any{
			"temperature": 0.15,
			"top_p":       0.9,
			"num_predict": 8192,
		}, collectChunk)
	}

	if err != nil {
		return accumulated.String(), err
	}

	res := accumulated.String()
	c.cache.Add(key, res)
	return res, nil
}

// streamGenerateCloud streams completions from any OpenAI-compatible endpoint (Gemini, Groq, OpenAI).
func (c *Client) streamGenerateCloud(
	ctx context.Context,
	systemPrompt string,
	userMessage string,
	onToken func(string) error,
) error {
	if c.apiKey == "" {
		return fmt.Errorf("API key is not configured for provider %s. Please set %s_API_KEY in backend/.env", c.provider, strings.ToUpper(c.provider))
	}

	reqBody := chatCompletionRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMessage},
		},
		Stream:      true,
		Temperature: 0.15,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal cloud request: %w", err)
	}

	url := c.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("create cloud request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("cloud request to %s failed: %w", c.provider, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("%s returned HTTP %d: %v", c.provider, resp.StatusCode, errResp)
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 64*1024), 64*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ":") {
			continue // SSE keepalive or comment
		}

		if !strings.HasPrefix(line, "data:") {
			continue
		}

		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			break
		}

		var chunk chatCompletionChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		if chunk.Error != nil && chunk.Error.Message != "" {
			return fmt.Errorf("%s error: %s", c.provider, chunk.Error.Message)
		}

		if len(chunk.Choices) > 0 {
			token := chunk.Choices[0].Delta.Content
			if token != "" {
				if err := onToken(token); err != nil {
					return err
				}
			}
		}
	}

	return scanner.Err()
}

// streamGenerateOllama streams from local Ollama /api/generate.
func (c *Client) streamGenerateOllama(
	ctx context.Context,
	prompt string,
	options map[string]any,
	onToken func(string) error,
) error {
	reqBody := ollamaGenerateRequest{
		Model:   c.model,
		Prompt:  prompt,
		Stream:  true,
		Options: options,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/api/generate", bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("ollama connection failed (is Ollama running on %s?): %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		return fmt.Errorf("ollama returned HTTP %d: %v", resp.StatusCode, errBody)
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 64*1024), 64*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var chunk ollamaGenerateChunk
		if err := json.Unmarshal(line, &chunk); err != nil {
			continue
		}

		if chunk.Error != "" {
			return fmt.Errorf("ollama error: %s", chunk.Error)
		}

		if chunk.Response != "" {
			if err := onToken(chunk.Response); err != nil {
				return fmt.Errorf("token callback failed: %w", err)
			}
		}

		if chunk.Done {
			break
		}
	}

	return scanner.Err()
}

func resolveOllamaModel(ollamaURL, preferred string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ollamaURL+"/api/tags", nil)
	if err != nil {
		return preferred
	}

	client := &http.Client{Timeout: 1 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return preferred
	}
	defer resp.Body.Close()

	var tags ollamaTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil || len(tags.Models) == 0 {
		return preferred
	}

	for _, m := range tags.Models {
		if m.Name == preferred || m.Name == preferred+":latest" {
			return m.Name
		}
	}

	basePref := preferred
	if idx := strings.IndexByte(preferred, ':'); idx != -1 {
		basePref = preferred[:idx]
	}
	for _, m := range tags.Models {
		if strings.HasPrefix(m.Name, basePref) {
			return m.Name
		}
	}

	return tags.Models[0].Name
}

func cacheKey(prompt string) string {
	h := sha256.Sum256([]byte(prompt))
	return fmt.Sprintf("%x", h[:8])
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

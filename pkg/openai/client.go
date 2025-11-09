package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// ClientConfig holds configuration for OpenAI client
type ClientConfig struct {
	APIKey      string        `json:"api_key"`
	Model       string        `json:"model"`
	BaseURL     string        `json:"base_url"`
	Timeout     time.Duration `json:"timeout"`
	MaxRetries  int           `json:"max_retries"`
	RetryDelay  time.Duration `json:"retry_delay"`
	UserAgent   string        `json:"user_agent"`
	MaxTokens   int           `json:"max_tokens"`
	Temperature float64       `json:"temperature"`
	Referer     string        `json:"referer"`
	AppTitle    string        `json:"app_title"`
}

// DefaultClientConfig returns sensible defaults
func DefaultClientConfig(apiKey string) ClientConfig {
	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-4o"
	}
	
	baseURL := os.Getenv("OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	
	// For OpenRouter, add openai/ prefix if not present
	if baseURL == "https://openrouter.ai/api/v1" && model == "gpt-4o" {
		model = "openai/gpt-4o"
	}
	
	referer := os.Getenv("OPENAI_REFERER")
	if referer == "" {
		referer = "https://kainuguru.com"
	}
	
	appTitle := os.Getenv("OPENAI_APP_TITLE")
	if appTitle == "" {
		appTitle = "Kainuguru"
	}
	
	return ClientConfig{
		APIKey:      apiKey,
		Model:       model,
		BaseURL:     baseURL,
		Timeout:     60 * time.Second,
		MaxRetries:  3,
		RetryDelay:  2 * time.Second,
		UserAgent:   "KainuguruBot/1.0",
		MaxTokens:   4000,
		Temperature: 0.1,
		Referer:     referer,
		AppTitle:    appTitle,
	}
}

// Client wraps OpenAI API interactions
type Client struct {
	config     ClientConfig
	httpClient *http.Client
}

// NewClient creates a new OpenAI client
func NewClient(config ClientConfig) *Client {
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// VisionRequest represents a request to OpenAI Vision API
type VisionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature,omitempty"`
}

// Message represents a message in the conversation
type Message struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

// Content represents content within a message
type Content struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

// ImageURL represents an image URL in the content
type ImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

// VisionResponse represents response from OpenAI Vision API
type VisionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice represents a choice in the response
type Choice struct {
	Index        int             `json:"index"`
	Message      ResponseMessage `json:"message"`
	FinishReason string          `json:"finish_reason"`
}

// ResponseMessage represents a message in the response
type ResponseMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ErrorResponse represents an error response from OpenAI
type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// AnalyzeImage sends an image to OpenAI Vision API for analysis
func (c *Client) AnalyzeImage(ctx context.Context, imageURL, prompt string) (*VisionResponse, error) {
	request := VisionRequest{
		Model: c.config.Model,
		Messages: []Message{
			{
				Role: "user",
				Content: []Content{
					{
						Type: "text",
						Text: prompt,
					},
					{
						Type: "image_url",
						ImageURL: &ImageURL{
							URL:    imageURL,
							Detail: "high",
						},
					},
				},
			},
		},
		MaxTokens:   c.config.MaxTokens,
		Temperature: c.config.Temperature,
	}

	return c.makeVisionRequest(ctx, request)
}

// AnalyzeImageWithBase64 sends a base64 encoded image to OpenAI Vision API
func (c *Client) AnalyzeImageWithBase64(ctx context.Context, base64Image, prompt string) (*VisionResponse, error) {
	request := VisionRequest{
		Model: c.config.Model,
		Messages: []Message{
			{
				Role: "user",
				Content: []Content{
					{
						Type: "text",
						Text: prompt,
					},
					{
						Type: "image_url",
						ImageURL: &ImageURL{
							URL:    base64Image, // Should include data:image/jpeg;base64, prefix
							Detail: "high",
						},
					},
				},
			},
		},
		MaxTokens:   c.config.MaxTokens,
		Temperature: c.config.Temperature,
	}

	return c.makeVisionRequest(ctx, request)
}

// BatchAnalyzeImages analyzes multiple images with the same prompt
func (c *Client) BatchAnalyzeImages(ctx context.Context, imageURLs []string, prompt string) ([]*VisionResponse, error) {
	responses := make([]*VisionResponse, len(imageURLs))

	for i, imageURL := range imageURLs {
		response, err := c.AnalyzeImage(ctx, imageURL, prompt)
		if err != nil {
			return responses, fmt.Errorf("failed to analyze image %d: %v", i, err)
		}
		responses[i] = response

		// Add delay between requests to respect rate limits
		if i < len(imageURLs)-1 {
			select {
			case <-ctx.Done():
				return responses, ctx.Err()
			case <-time.After(c.config.RetryDelay):
			}
		}
	}

	return responses, nil
}

// makeVisionRequest makes the actual HTTP request to OpenAI Vision API
func (c *Client) makeVisionRequest(ctx context.Context, request VisionRequest) (*VisionResponse, error) {
	// Serialize request
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create HTTP request
	url := c.config.BaseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	httpReq.Header.Set("User-Agent", c.config.UserAgent)
	
	// Add OpenRouter specific headers if configured
	if c.config.Referer != "" {
		httpReq.Header.Set("HTTP-Referer", c.config.Referer)
	}
	if c.config.AppTitle != "" {
		httpReq.Header.Set("X-Title", c.config.AppTitle)
	}

	// Make request with retries
	var response *VisionResponse
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			if attempt == c.config.MaxRetries {
				return nil, fmt.Errorf("request failed after %d attempts: %v", c.config.MaxRetries+1, err)
			}
			time.Sleep(c.config.RetryDelay * time.Duration(attempt+1))
			continue
		}

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %v", err)
		}

		// Handle different status codes
		switch resp.StatusCode {
		case http.StatusOK:
			if err := json.Unmarshal(body, &response); err != nil {
				return nil, fmt.Errorf("failed to unmarshal response: %v (body preview: %s)", err, string(body[:min(200, len(body))]))
			}
			return response, nil

		case http.StatusTooManyRequests:
			if attempt == c.config.MaxRetries {
				return nil, fmt.Errorf("rate limited after %d attempts", c.config.MaxRetries+1)
			}
			// Exponential backoff for rate limiting
			delay := c.config.RetryDelay * time.Duration(1<<attempt)
			time.Sleep(delay)
			continue

		case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden:
			var errorResp ErrorResponse
			if err := json.Unmarshal(body, &errorResp); err == nil {
				return nil, fmt.Errorf("API error: %s", errorResp.Error.Message)
			}
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))

		default:
			return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
		}
	}

	return response, nil
}

// ExtractText analyzes an image and extracts text content
func (c *Client) ExtractText(ctx context.Context, imageURL string) (string, error) {
	prompt := `Extract all text visible in this image. Include product names, prices, descriptions, and any other text.
Format the output as clean, readable text. Preserve the original language (Lithuanian if present).`

	response, err := c.AnalyzeImage(ctx, imageURL, prompt)
	if err != nil {
		return "", err
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	return response.Choices[0].Message.Content, nil
}

// ExtractProducts analyzes a grocery flyer page and extracts product information
func (c *Client) ExtractProducts(ctx context.Context, imageURL string) (string, error) {
	prompt := `Analyze this grocery store flyer page and extract all visible products. For each product, provide:
1. Product name (in original language, likely Lithuanian)
2. Price (including currency)
3. Unit/quantity information
4. Any discount/sale information
5. Brand name if visible

Format as JSON array with objects containing: name, price, unit, discount, brand fields.
Only include products that are clearly visible with prices.`

	response, err := c.AnalyzeImage(ctx, imageURL, prompt)
	if err != nil {
		return "", err
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	return response.Choices[0].Message.Content, nil
}

// GetTokenUsage returns the total tokens used in a response
func (r *VisionResponse) GetTokenUsage() int {
	return r.Usage.TotalTokens
}

// GetContent returns the content of the first choice
func (r *VisionResponse) GetContent() string {
	if len(r.Choices) > 0 {
		return r.Choices[0].Message.Content
	}
	return ""
}

// IsComplete checks if the response was completed successfully
func (r *VisionResponse) IsComplete() bool {
	if len(r.Choices) > 0 {
		return r.Choices[0].FinishReason == "stop"
	}
	return false
}

// EstimateTokens provides a rough estimate of tokens for text
func EstimateTokens(text string) int {
	// Rough approximation: 1 token ≈ 4 characters for English
	// This is a simplified estimate; actual tokenization is more complex
	return len(text) / 4
}

// ValidateAPIKey checks if the API key format is valid
func ValidateAPIKey(apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("API key is empty")
	}

	if len(apiKey) < 40 {
		return fmt.Errorf("API key is too short")
	}

	if !bytes.HasPrefix([]byte(apiKey), []byte("sk-")) {
		return fmt.Errorf("API key should start with 'sk-'")
	}

	return nil
}

// GetModelLimits returns the known limits for different models
func GetModelLimits(model string) ModelLimits {
	limits := map[string]ModelLimits{
		"gpt-4o": {
			MaxTokens:       4096,
			MaxImageSize:    20 * 1024 * 1024, // 20MB
			MaxImages:       10,
			CostPer1KTokens: 0.01,
		},
		"gpt-4-vision-preview": {  // Deprecated but keeping for reference
			MaxTokens:       4096,
			MaxImageSize:    20 * 1024 * 1024,
			MaxImages:       10,
			CostPer1KTokens: 0.01,
		},
		"gpt-4": {
			MaxTokens:       8192,
			MaxImageSize:    0, // No image support
			MaxImages:       0,
			CostPer1KTokens: 0.03,
		},
	}

	if limit, exists := limits[model]; exists {
		return limit
	}

	// Default limits for gpt-4o
	return ModelLimits{
		MaxTokens:       4096,
		MaxImageSize:    20 * 1024 * 1024,
		MaxImages:       10,
		CostPer1KTokens: 0.01,
	}
}

// ModelLimits represents the limits for a specific model
type ModelLimits struct {
	MaxTokens       int     `json:"max_tokens"`
	MaxImageSize    int64   `json:"max_image_size"`
	MaxImages       int     `json:"max_images"`
	CostPer1KTokens float64 `json:"cost_per_1k_tokens"`
}

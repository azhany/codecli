package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/azhany/codecli/internal/config"
)

// Client represents the LLM client
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new LLM client
func NewClient() (*Client, error) {
	cfg := config.Config.Ollama
	return &Client{
		httpClient: &http.Client{},
		baseURL:    cfg.URL,
	}, nil
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type ChatResponse struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

type EmbeddingsRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type EmbeddingsResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

// Chat sends a message to the LLM and processes the response
func (c *Client) Chat(ctx context.Context, message string, tools []string) (string, error) {
	reqBody := ChatRequest{
		Model: config.Config.Ollama.ChatModel,
		Messages: []Message{
			{Role: "user", Content: message},
		},
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/chat", bytes.NewReader(reqBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	return chatResp.Message.Content, nil
}

// EmbedText generates embeddings for text
func (c *Client) EmbedText(ctx context.Context, text string) ([]float32, error) {
	reqBody := EmbeddingsRequest{
		Model:  config.Config.Ollama.EmbeddingModel,
		Prompt: text,
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/embeddings", bytes.NewReader(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var embedResp EmbeddingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if len(embedResp.Embeddings) == 0 || len(embedResp.Embeddings[0]) == 0 {
		return nil, fmt.Errorf("empty embeddings in response")
	}

	return embedResp.Embeddings[0], nil
}

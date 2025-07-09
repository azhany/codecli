package llm

import (
	"context"
	"fmt"

	"github.com/azhany/codecli/internal/config"
	"github.com/azhany/codecli/internal/tools"
	"github.com/ollama/go-ollama"
)

// Client represents the LLM client
type Client struct {
	client *ollama.Client
	tools  map[string]tools.Tool
}

// NewClient creates a new LLM client
func NewClient() (*Client, error) {
	cfg := config.Config.Ollama
	client := ollama.NewClient(cfg.URL)

	// Initialize tools
	tools := make(map[string]tools.Tool)
	for toolType, toolFactory := range tools.ToolRegistry {
		tool := toolFactory()
		tools[tool.Name()] = tool
	}

	return &Client{
		client: client,
		tools:  tools,
	}, nil
}

// Chat sends a message to the LLM and processes the response
func (c *Client) Chat(ctx context.Context, message string, tools []string) (string, error) {
	// Create chat request
	req := ollama.ChatRequest{
		Model: config.Config.Ollama.ChatModel,
		Messages: []ollama.Message{
			{Role: "user", Content: message},
		},
		Tools: tools,
	}

	// Send request
	stream, err := c.client.Chat(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to start chat: %v", err)
	}

	// Process response
	var response string
	for {
		msg, err := stream.Recv()
		if err != nil {
			break
		}
		response += msg.Content
	}

	return response, nil
}

// ExecuteTool executes a tool with the given arguments
func (c *Client) ExecuteTool(toolName string, args map[string]interface{}) (interface{}, error) {
	tool, exists := c.tools[toolName]
	if !exists {
		return nil, fmt.Errorf("tool %s not found", toolName)
	}

	return tool.Execute(args)
}

// EmbedText generates embeddings for text
func (c *Client) EmbedText(ctx context.Context, text string) ([]float32, error) {
	req := ollama.EmbedRequest{
		Model: config.Config.Ollama.EmbeddingModel,
		Text:  text,
	}

	resp, err := c.client.Embed(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %v", err)
	}

	return resp.Embedding, nil
}

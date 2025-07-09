package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/azhany/codecli/internal/config"
	"github.com/azhany/codecli/internal/llm"
	"github.com/azhany/codecli/internal/vector"
)

// Tool represents a tool that can be called by the LLM
//
//go:generate stringer -type=ToolType
//go:generate stringer -type=FileType
type ToolType int

type FileType int

const (
	ExecuteCommand ToolType = iota
	ReadFile
	WriteToFile
	ListFiles
	ListCodeDefinitions
	SearchFiles
	AskFollowup
)

const (
	GoFileType FileType = iota
	PythonFileType
	JavaScriptFileType
	TypeScriptFileType
	JavaFileType
	CppFileType
	CFileType
	HFileType
)

// ToolRegistry holds all available tools
var ToolRegistry = map[ToolType]func() Tool{
	ExecuteCommand:      NewCommandTool,
	ReadFile:            NewFileReader,
	WriteToFile:         NewFileWriter,
	ListFiles:           NewFileLister,
	ListCodeDefinitions: NewCodeDefinitionLister,
	SearchFiles:         NewFileSearcher,
	AskFollowup:         NewFollowupTool,
}

// Tool interface that all tools must implement
type Tool interface {
	Name() string
	Description() string
	Execute(args map[string]interface{}) (interface{}, error)
}

// IndexCodebase indexes the codebase for semantic search
func IndexCodebase() error {
	// Get workspace configuration
	ws := config.Config.Workspace

	// Create vector store instance
	store, err := vector.NewVectorStore()
	if err != nil {
		return fmt.Errorf("failed to create vector store: %v", err)
	}
	defer store.Close()

	// Create vector index
	if err := store.CreateIndex(ws.Root, ws.IncludeExtensions); err != nil {
		return fmt.Errorf("failed to create vector index: %v", err)
	}

	return nil
}

// SearchCodebase performs a semantic search on the codebase
func SearchCodebase(query string, limit int) ([]vector.SearchResult, error) {
	// Create vector store instance
	store, err := vector.NewVectorStore()
	if err != nil {
		return nil, fmt.Errorf("failed to create vector store: %v", err)
	}
	defer store.Close()

	// Try to load existing index
	if err := store.LoadIndex(); err != nil {
		return nil, fmt.Errorf("failed to load index: %v. Please run index command first", err)
	}

	// Perform search
	results, err := store.Search(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %v", err)
	}

	return results, nil
}

// StartChat starts the interactive chat mode
func StartChat() error {
	llmClient, err := llm.NewClient()
	if err != nil {
		return fmt.Errorf("failed to initialize LLM client: %v", err)
	}

	// Create vector store for semantic search
	store, err := vector.NewVectorStore()
	if err != nil {
		return fmt.Errorf("failed to create vector store: %v", err)
	}
	defer store.Close()

	// Try to load existing index (optional)
	indexLoaded := false
	if err := store.LoadIndex(); err == nil {
		indexLoaded = true
		fmt.Println("✓ Loaded existing codebase index")
	} else {
		fmt.Println("⚠ No existing index found. Some features may be limited.")
		fmt.Println("  Run 'codecli index' to enable semantic search.")
	}

	fmt.Println("🤖 CodeCLI Chat Mode")
	fmt.Println("Type 'exit' or 'quit' to end the session")
	fmt.Println("Type '/search <query>' to search the codebase")
	fmt.Println("Type '/help' for more commands")
	fmt.Println(strings.Repeat("-", 50))

	scanner := bufio.NewScanner(os.Stdin)
	ctx := context.Background()

	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// Handle special commands
		if input == "exit" || input == "quit" {
			fmt.Println("👋 Goodbye!")
			break
		}

		if input == "/help" {
			printHelpMessage()
			continue
		}

		if strings.HasPrefix(input, "/search ") {
			query := strings.TrimPrefix(input, "/search ")
			if !indexLoaded {
				fmt.Println("❌ Search requires an indexed codebase. Run 'codecli index' first.")
				continue
			}

			results, err := store.Search(query, 5)
			if err != nil {
				fmt.Printf("❌ Search failed: %v\n", err)
				continue
			}

			if len(results) == 0 {
				fmt.Println("🔍 No results found")
				continue
			}

			fmt.Printf("🔍 Found %d results:\n", len(results))
			for i, result := range results {
				fmt.Printf("\n%d. %s\n", i+1, result.String())
			}
			continue
		}

		// Send message to LLM
		fmt.Print("🤖 Assistant: ")
		response, err := llmClient.Chat(ctx, input, []string{})
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}

		fmt.Println(response)
		fmt.Println()
	}

	return scanner.Err()
}

// printHelpMessage prints available commands
func printHelpMessage() {
	fmt.Println("Available commands:")
	fmt.Println("  /help           - Show this help message")
	fmt.Println("  /search <query> - Search the codebase semantically")
	fmt.Println("  exit, quit      - Exit the chat session")
	fmt.Println()
	fmt.Println("You can also ask questions directly and the AI will respond.")
}

// ListFiles lists files in the workspace
func ListFiles(root string, pattern string) ([]string, error) {
	var files []string

	// Walk directory and filter files
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && matchesPattern(path, pattern) {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// matchesPattern checks if a file matches the given pattern
func matchesPattern(path string, pattern string) bool {
	if pattern == "" || pattern == "*" {
		return true
	}

	// Get the filename from the path
	filename := filepath.Base(path)

	// Handle simple wildcard patterns
	if strings.Contains(pattern, "*") {
		matched, err := filepath.Match(pattern, filename)
		if err != nil {
			// If pattern is invalid, fall back to substring matching
			return strings.Contains(strings.ToLower(filename), strings.ToLower(pattern))
		}
		return matched
	}

	// Handle extension matching (e.g., ".go", ".py")
	if strings.HasPrefix(pattern, ".") {
		return strings.HasSuffix(strings.ToLower(path), strings.ToLower(pattern))
	}

	// Handle directory patterns (e.g., "src/", "/test/")
	if strings.Contains(pattern, "/") {
		return strings.Contains(strings.ToLower(path), strings.ToLower(pattern))
	}

	// Default: substring matching (case-insensitive)
	return strings.Contains(strings.ToLower(filename), strings.ToLower(pattern))
}

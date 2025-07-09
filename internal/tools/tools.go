package tools

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/azhany/codecli/internal/config"
	"github.com/azhany/codecli/internal/llm"
	"github.com/azhany/codecli/internal/vector"
)

// Tool represents a tool that can be called by the LLM
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
	ExecuteCommand: NewCommandTool,
	ReadFile:       NewFileReader,
	WriteToFile:    NewFileWriter,
	ListFiles:      NewFileLister,
	ListCodeDefinitions: NewCodeDefinitionLister,
	SearchFiles:    NewFileSearcher,
	AskFollowup:   NewFollowupTool,
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
	
	// Create vector index
	if err := vector.CreateIndex(ws.Root, ws.IncludeExtensions); err != nil {
		return fmt.Errorf("failed to create vector index: %v", err)
	}
	
	return nil
}

// SearchCodebase performs a semantic search on the codebase
func SearchCodebase() error {
	// TODO: Implement search functionality
	return nil
}

// StartChat starts the interactive chat mode
func StartChat() error {
	llmClient, err := llm.NewClient()
	if err != nil {
		return fmt.Errorf("failed to initialize LLM client: %v", err)
	}
	
	// TODO: Implement chat loop
	return nil
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
	// TODO: Implement pattern matching logic
	return true
}

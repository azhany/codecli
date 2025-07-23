// Package types provides core interfaces and types for the CLI tools
package types

// Tool represents a tool that can be called by the LLM
type Tool interface {
    Name() string
    Description() string
    Execute(args map[string]interface{}) (interface{}, error)
}

// FileHandler handles file operations like read, write, list, and search
type FileHandler interface {
    Tool
    HandleFile(operation string, path string, data []byte) ([]byte, error)
}

// CommandRunner handles command execution
type CommandRunner interface {
    Tool
    RunCommand(cmd string, args ...string) (string, error)
}

// SearchResult represents a single search result
type SearchResult struct {
    Path     string
    Line     int
    Content  string
    Distance float64
}

// ToolFactory creates tool instances
type ToolFactory func() Tool

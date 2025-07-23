package tools

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/azhany/codecli/internal/search"
	"github.com/azhany/codecli/internal/types"
)

type FileOperation string

const (
	FileRead   FileOperation = "read"
	FileWrite  FileOperation = "write"
	FileList   FileOperation = "list"
	FileSearch FileOperation = "search"
)

// File handles file operations
type File struct {
	*Base
}

func NewFile() *File {
	return &File{
		Base: NewBase("file", "Handles file operations (read/write/list/search)"),
	}
}

func (t *File) HandleFile(operation string, path string, data []byte) ([]byte, error) {
	switch FileOperation(operation) {
	case FileRead:
		return os.ReadFile(path)
	case FileWrite:
		return nil, os.WriteFile(path, data, 0644)
	case FileList:
		files, err := t.listFiles(path, string(data))
		if err != nil {
			return nil, err
		}
		result := []byte(fmt.Sprintf("%v", files))
		return result, nil
	case FileSearch:
		results, err := t.searchFiles(string(data), 10)
		if err != nil {
			return nil, err
		}
		result := []byte(fmt.Sprintf("%v", results))
		return result, nil
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (t *File) listFiles(root string, pattern string) ([]string, error) {
	if root == "" {
		root = "."
	}
	if pattern == "" {
		pattern = "*"
	}

	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %v", err)
	}

	return files, nil
}

func (t *File) searchFiles(query string, limit int) ([]types.SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	return search.SearchCodebase(query, limit)
}

func (t *File) Execute(args map[string]interface{}) (interface{}, error) {
	operation, ok := args["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("operation argument is required")
	}

	path, ok := args["path"].(string)
	if !ok {
		path = "."
	}

	var data []byte
	if content, ok := args["content"].(string); ok {
		data = []byte(content)
	}

	result, err := t.HandleFile(operation, path, data)
	if err != nil {
		return nil, err
	}

	if op := FileOperation(operation); op == FileRead || op == FileList || op == FileSearch {
		return string(result), nil
	}
	return nil, nil
}

package vector

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nipponnagoya/ngt-go"
	"github.com/azhany/codecli/internal/config"
)

// VectorStore represents the NGT vector store
type VectorStore struct {
	index *ngt.Index
	config *ngt.Config
}

// NewVectorStore creates a new vector store
func NewVectorStore() (*VectorStore, error) {
	cfg := &ngt.Config{
		Dimension: config.Config.NGT.Dimension,
		EdgeSize:  config.Config.NGT.EdgeSize,
	}
	
	store := &VectorStore{
		config: cfg,
	}
	
	// Create index directory if it doesn't exist
	indexPath := config.Config.NGT.IndexPath
	if err := os.MkdirAll(indexPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create index directory: %v", err)
	}
	
	return store, nil
}

// CreateIndex creates a new vector index for the codebase
func (v *VectorStore) CreateIndex(root string, extensions []string) error {
	// Initialize NGT
	index, err := ngt.NewIndex(v.config)
	if err != nil {
		return fmt.Errorf("failed to initialize NGT: %v", err)
	}
	v.index = index
	
	// Process files
	files, err := findCodeFiles(root, extensions)
	if err != nil {
		return fmt.Errorf("failed to find code files: %v", err)
	}
	
	for _, file := range files {
		if err := v.processFile(file); err != nil {
			return fmt.Errorf("failed to process file %s: %v", file, err)
		}
	}
	
	return nil
}

// Search performs a semantic search on the codebase
func (v *VectorStore) Search(query string, limit int) ([]string, error) {
	// TODO: Implement search functionality
	return nil, nil
}

// Close closes the vector store
func (v *VectorStore) Close() error {
	if v.index != nil {
		return v.index.Close()
	}
	return nil
}

// findCodeFiles finds code files in the workspace
func findCodeFiles(root string, extensions []string) ([]string, error) {
	var files []string
	
	// Walk directory and filter files
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() {
			for _, ext := range extensions {
				if filepath.Ext(path) == ext {
					files = append(files, path)
					break
				}
			}
		}
		
		return nil
	})
	
	return files, err
}

// processFile processes a single file and adds its vectors to the index
func (v *VectorStore) processFile(file string) error {
	// TODO: Implement file processing and vector generation
	return nil
}

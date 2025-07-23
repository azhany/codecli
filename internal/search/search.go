package search

import (
	"github.com/azhany/codecli/internal/types"
)

// Engine represents a code search engine
type Engine interface {
	Search(query string, limit int) ([]types.SearchResult, error)
}

// DefaultEngine is the default implementation of the search engine
type DefaultEngine struct {
	// Add fields for vector store, etc.
}

// NewDefaultEngine creates a new default search engine
func NewDefaultEngine() *DefaultEngine {
	return &DefaultEngine{}
}

// Search performs a semantic search using the vector store
func (e *DefaultEngine) Search(query string, limit int) ([]types.SearchResult, error) {
	// TODO: Implement semantic search using vector store
	return []types.SearchResult{}, nil
}

// SearchCodebase is a convenience function that uses the default engine
func SearchCodebase(query string, limit int) ([]types.SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	engine := NewDefaultEngine()
	return engine.Search(query, limit)
}

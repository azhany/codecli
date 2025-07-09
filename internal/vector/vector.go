package vector

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/azhany/codecli/internal/config"
	"github.com/azhany/codecli/internal/llm"
	"github.com/nipponnagoya/ngt-go"
)

// FileMetadata represents metadata for indexed files
type FileMetadata struct {
	ID       uint32
	FilePath string
	Content  string
	Chunks   []ChunkMetadata
}

// ChunkMetadata represents metadata for file chunks
type ChunkMetadata struct {
	ID        uint32
	StartLine int
	EndLine   int
	Content   string
}

// VectorStore represents the NGT vector store
type VectorStore struct {
	index     *ngt.Index
	config    *ngt.Config
	llmClient *llm.Client
	metadata  map[uint32]*FileMetadata
	mutex     sync.RWMutex
	nextID    uint32
}

// NewVectorStore creates a new vector store
func NewVectorStore() (*VectorStore, error) {
	cfg := &ngt.Config{
		Dimension: config.Config.NGT.Dimension,
		EdgeSize:  config.Config.NGT.EdgeSize,
	}

	// Initialize LLM client
	llmClient, err := llm.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize LLM client: %v", err)
	}

	store := &VectorStore{
		config:    cfg,
		llmClient: llmClient,
		metadata:  make(map[uint32]*FileMetadata),
		nextID:    1,
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

	// Save index and metadata to disk
	if err := v.saveIndex(); err != nil {
		return fmt.Errorf("failed to save index: %v", err)
	}

	return nil
}

// SearchResult represents a search result
type SearchResult struct {
	FilePath  string
	Content   string
	StartLine int
	EndLine   int
	Score     float64
}

// Search performs a semantic search on the codebase
func (v *VectorStore) Search(query string, limit int) ([]SearchResult, error) {
	if v.index == nil {
		return nil, fmt.Errorf("index not initialized")
	}

	// Generate embedding for query
	ctx := context.Background()
	queryEmbedding, err := v.llmClient.EmbedText(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %v", err)
	}

	// Convert float32 to float64 for NGT
	queryVector := make([]float64, len(queryEmbedding))
	for i, val := range queryEmbedding {
		queryVector[i] = float64(val)
	}

	// Search in NGT index
	results, err := v.index.Search(queryVector, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search index: %v", err)
	}

	// Convert results to SearchResult format
	searchResults := make([]SearchResult, 0, len(results))

	v.mutex.RLock()
	defer v.mutex.RUnlock()

	for _, result := range results {
		// Find the chunk metadata
		var foundChunk *ChunkMetadata
		var foundFile *FileMetadata

		for _, fileMeta := range v.metadata {
			for _, chunk := range fileMeta.Chunks {
				if chunk.ID == result.ID {
					foundChunk = &chunk
					foundFile = fileMeta
					break
				}
			}
			if foundChunk != nil {
				break
			}
		}

		if foundChunk != nil && foundFile != nil {
			searchResult := SearchResult{
				FilePath:  foundFile.FilePath,
				Content:   foundChunk.Content,
				StartLine: foundChunk.StartLine,
				EndLine:   foundChunk.EndLine,
				Score:     result.Distance, // NGT returns distance, lower is better
			}
			searchResults = append(searchResults, searchResult)
		}
	}

	return searchResults, nil
}

// saveIndex saves the NGT index and metadata to disk
func (v *VectorStore) saveIndex() error {
	indexPath := config.Config.NGT.IndexPath

	// Save NGT index
	if err := v.index.Save(indexPath); err != nil {
		return fmt.Errorf("failed to save NGT index: %v", err)
	}

	// Save metadata
	metadataPath := filepath.Join(indexPath, "metadata.json")
	v.mutex.RLock()
	metadataBytes, err := json.Marshal(v.metadata)
	v.mutex.RUnlock()

	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %v", err)
	}

	if err := ioutil.WriteFile(metadataPath, metadataBytes, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %v", err)
	}

	return nil
}

// LoadIndex loads an existing index from disk
func (v *VectorStore) LoadIndex() error {
	indexPath := config.Config.NGT.IndexPath

	// Check if index exists
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		return fmt.Errorf("index does not exist at path: %s", indexPath)
	}

	// Load NGT index
	index, err := ngt.LoadIndex(indexPath)
	if err != nil {
		return fmt.Errorf("failed to load NGT index: %v", err)
	}
	v.index = index

	// Load metadata
	metadataPath := filepath.Join(indexPath, "metadata.json")
	if _, err := os.Stat(metadataPath); err == nil {
		metadataBytes, err := ioutil.ReadFile(metadataPath)
		if err != nil {
			return fmt.Errorf("failed to read metadata file: %v", err)
		}

		v.mutex.Lock()
		if err := json.Unmarshal(metadataBytes, &v.metadata); err != nil {
			v.mutex.Unlock()
			return fmt.Errorf("failed to unmarshal metadata: %v", err)
		}

		// Find the highest ID to set nextID
		maxID := uint32(0)
		for _, fileMeta := range v.metadata {
			if fileMeta.ID > maxID {
				maxID = fileMeta.ID
			}
			for _, chunk := range fileMeta.Chunks {
				if chunk.ID > maxID {
					maxID = chunk.ID
				}
			}
		}
		v.nextID = maxID + 1
		v.mutex.Unlock()
	}

	return nil
}

// Close closes the vector store
func (v *VectorStore) Close() error {
	if v.index != nil {
		return v.index.Close()
	}
	return nil
}

// FormatSearchResults formats search results for display
func (sr SearchResult) String() string {
	return fmt.Sprintf("File: %s (lines %d-%d, score: %.4f)\n%s",
		sr.FilePath, sr.StartLine, sr.EndLine, sr.Score, sr.Content)
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
	// Read file content
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	// Split content into chunks
	chunks := v.splitIntoChunks(string(content))
	if len(chunks) == 0 {
		return nil // Skip empty files
	}

	v.mutex.Lock()
	fileID := v.nextID
	v.nextID++
	v.mutex.Unlock()

	// Create file metadata
	fileMeta := &FileMetadata{
		ID:       fileID,
		FilePath: file,
		Content:  string(content),
		Chunks:   make([]ChunkMetadata, 0, len(chunks)),
	}

	// Process each chunk
	for i, chunk := range chunks {
		if strings.TrimSpace(chunk.Content) == "" {
			continue // Skip empty chunks
		}

		// Generate embedding for chunk
		ctx := context.Background()
		embedding, err := v.llmClient.EmbedText(ctx, chunk.Content)
		if err != nil {
			return fmt.Errorf("failed to generate embedding for chunk: %v", err)
		}

		// Convert float32 to float64 for NGT
		vector := make([]float64, len(embedding))
		for j, val := range embedding {
			vector[j] = float64(val)
		}

		v.mutex.Lock()
		chunkID := v.nextID
		v.nextID++
		v.mutex.Unlock()

		// Add vector to index
		if err := v.index.Insert(chunkID, vector); err != nil {
			return fmt.Errorf("failed to insert vector: %v", err)
		}

		// Add chunk metadata
		chunk.ID = chunkID
		fileMeta.Chunks = append(fileMeta.Chunks, chunk)
	}

	// Store file metadata
	v.mutex.Lock()
	v.metadata[fileID] = fileMeta
	v.mutex.Unlock()

	return nil
}

// splitIntoChunks splits file content into manageable chunks
func (v *VectorStore) splitIntoChunks(content string) []ChunkMetadata {
	lines := strings.Split(content, "\n")
	chunks := make([]ChunkMetadata, 0)

	const maxLinesPerChunk = 50 // Configurable chunk size
	const overlapLines = 5      // Lines to overlap between chunks

	for i := 0; i < len(lines); i += maxLinesPerChunk - overlapLines {
		endIdx := i + maxLinesPerChunk
		if endIdx > len(lines) {
			endIdx = len(lines)
		}

		chunkLines := lines[i:endIdx]
		chunkContent := strings.Join(chunkLines, "\n")

		if strings.TrimSpace(chunkContent) != "" {
			chunk := ChunkMetadata{
				StartLine: i + 1, // 1-based line numbering
				EndLine:   endIdx,
				Content:   chunkContent,
			}
			chunks = append(chunks, chunk)
		}

		// Break if we've reached the end
		if endIdx >= len(lines) {
			break
		}
	}

	return chunks
}

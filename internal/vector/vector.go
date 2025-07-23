package vector

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/azhany/codecli/internal/config"
	"github.com/azhany/codecli/internal/llm"
	"github.com/azhany/codecli/internal/types"
)

// Store handles vector storage and retrieval
type Store struct {
	embeddings map[string][]float64
	metadata   map[string]*FileMetadata
	mu         sync.RWMutex
	llmClient  *llm.Client
}

// NewStore creates a new vector store
func NewStore() (*Store, error) {
	client, err := llm.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client: %w", err)
	}

	return &Store{
		embeddings: make(map[string][]float64),
		metadata:   make(map[string]*FileMetadata),
		llmClient:  client,
	}, nil
}

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

// ChunkVector represents a chunk with its embedding vector
type ChunkVector struct {
	ChunkMetadata
	Vector []float32
}

// VectorStore represents the in-memory vector store
type VectorStore struct {
	llmClient *llm.Client
	metadata  map[uint32]*FileMetadata
	vectors   map[uint32]*ChunkVector // Map of chunk ID to vector
	mutex     sync.RWMutex
	nextID    uint32
}

// NewVectorStore creates a new vector store
func NewVectorStore() (*VectorStore, error) {
	// Initialize LLM client
	llmClient, err := llm.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize LLM client: %v", err)
	}

	store := &VectorStore{
		llmClient: llmClient,
		metadata:  make(map[uint32]*FileMetadata),
		vectors:   make(map[uint32]*ChunkVector),
		nextID:    1,
	}

	// Create index directory if it doesn't exist
	indexPath := config.Config.NGT.IndexPath
	if err := os.MkdirAll(indexPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create index directory: %v", err)
	}

	return store, nil
}

// cosineSimilarity calculates the cosine similarity between two vectors
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct float64
	var normA float64
	var normB float64

	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i] * b[i])
		normA += float64(a[i] * a[i])
		normB += float64(b[i] * b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// CreateIndex creates a new vector index for the codebase
func (v *VectorStore) CreateIndex(root string, extensions []string) error {
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

	// Save metadata to disk
	if err := v.saveIndex(); err != nil {
		return fmt.Errorf("failed to save index: %v", err)
	}

	return nil
}

// Search performs a semantic search on the codebase
func (v *VectorStore) Search(query string, limit int) ([]types.SearchResult, error) {
	// Generate embedding for query
	ctx := context.Background()
	queryEmbedding, err := v.llmClient.EmbedText(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %v", err)
	}

	type scoreEntry struct {
		chunkVec *ChunkVector
		fileMeta *FileMetadata
		score    float64
	}

	// Calculate cosine similarity for all vectors
	var scores []scoreEntry

	v.mutex.RLock()
	for _, fileMeta := range v.metadata {
		for _, chunk := range fileMeta.Chunks {
			if vec, ok := v.vectors[chunk.ID]; ok {
				score := cosineSimilarity(queryEmbedding, vec.Vector)
				scores = append(scores, scoreEntry{
					chunkVec: vec,
					fileMeta: fileMeta,
					score:    score,
				})
			}
		}
	}
	v.mutex.RUnlock()

	// Sort by score (higher is better for cosine similarity)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// Take top K results
	if limit > len(scores) {
		limit = len(scores)
	}

	// Convert to SearchResult format
	searchResults := make([]types.SearchResult, 0, limit)
	for i := 0; i < limit; i++ {
		result := scores[i]
		searchResults = append(searchResults, types.SearchResult{
			Path:     result.fileMeta.FilePath,
			Line:     result.chunkVec.StartLine,
			Content:  result.chunkVec.Content,
			Distance: result.score,
		})
	}

	return searchResults, nil
}

// saveIndex saves metadata and vectors to disk
func (v *VectorStore) saveIndex() error {
	indexPath := config.Config.NGT.IndexPath

	v.mutex.RLock()
	data := struct {
		Metadata map[uint32]*FileMetadata `json:"metadata"`
		Vectors  map[uint32]*ChunkVector  `json:"vectors"`
	}{
		Metadata: v.metadata,
		Vectors:  v.vectors,
	}
	metadataBytes, err := json.Marshal(data)
	v.mutex.RUnlock()

	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}

	metadataPath := filepath.Join(indexPath, "metadata.json")
	if err := ioutil.WriteFile(metadataPath, metadataBytes, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %v", err)
	}

	return nil
}

// LoadIndex loads metadata and vectors from disk
func (v *VectorStore) LoadIndex() error {
	indexPath := config.Config.NGT.IndexPath
	metadataPath := filepath.Join(indexPath, "metadata.json")

	// Check if metadata exists
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		return fmt.Errorf("index does not exist at path: %s", metadataPath)
	}

	// Load metadata and vectors
	metadataBytes, err := ioutil.ReadFile(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to read metadata file: %v", err)
	}

	var data struct {
		Metadata map[uint32]*FileMetadata `json:"metadata"`
		Vectors  map[uint32]*ChunkVector  `json:"vectors"`
	}

	if err := json.Unmarshal(metadataBytes, &data); err != nil {
		return fmt.Errorf("failed to unmarshal data: %v", err)
	}

	v.mutex.Lock()
	v.metadata = data.Metadata
	v.vectors = data.Vectors

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

	return nil
}

// Close cleans up resources
func (v *VectorStore) Close() error {
	return nil
}

// FormatSearchResult formats a search result for display
func FormatSearchResult(sr types.SearchResult) string {
	return fmt.Sprintf("File: %s (line %d, score: %.4f)\n%s",
		sr.Path, sr.Line, sr.Distance, sr.Content)
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

// processFile processes a single file and adds its vectors to the store
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
	for _, chunk := range chunks {
		if strings.TrimSpace(chunk.Content) == "" {
			continue // Skip empty chunks
		}

		// Generate embedding for chunk
		ctx := context.Background()
		embedding, err := v.llmClient.EmbedText(ctx, chunk.Content)
		if err != nil {
			return fmt.Errorf("failed to generate embedding for chunk: %v", err)
		}

		v.mutex.Lock()
		chunkID := v.nextID
		v.nextID++

		// Store chunk vector
		chunkVec := &ChunkVector{
			ChunkMetadata: chunk,
			Vector:        embedding,
		}
		chunkVec.ID = chunkID
		v.vectors[chunkID] = chunkVec

		// Add chunk metadata
		fileMeta.Chunks = append(fileMeta.Chunks, chunk)
		v.mutex.Unlock()
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

package chunker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ChunkFile is the standard JSON envelope for serialised chunk data.
// It includes document-level metadata and processing settings.
type ChunkFile struct {
	Source      string    `json:"Source"`
	DocumentID  string    `json:"DocumentID"`
	TotalChunks int       `json:"TotalChunks"`
	ChunkSize   int       `json:"ChunkSize"`
	Overlap     int       `json:"Overlap"`
	CreatedAt   time.Time `json:"CreatedAt"`
	Chunks      []Chunk   `json:"Chunks"`
}

// WriteJSON serialises chunks to a JSON file inside outputDir.
//
// The output filename is derived from the source document name so files
// are easy to correlate:
//
//	"reports/q3-2024.txt"  →  "<outputDir>/q3-2024_chunks.json"
//
// The output directory is created automatically if it doesn't exist.
// Existing files with the same name are overwritten.
func WriteJSON(
	chunks []Chunk,
	source, documentID,
	outputDir string,
	opts Options) (string, error) {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", fmt.Errorf("chunker: create output dir %q: %w", outputDir, err)
	}

	filename := chunkFileName(source)
	outPath := filepath.Join(outputDir, filename)

	envelope := ChunkFile{
		Source:      source,
		DocumentID:  documentID,
		TotalChunks: len(chunks),
		ChunkSize:   opts.Size,
		Overlap:     opts.Overlap,
		CreatedAt:   time.Now(),
		Chunks:      chunks,
	}

	f, err := os.Create(outPath)
	if err != nil {
		return "", fmt.Errorf("chunker: create file %q: %w", outPath, err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(envelope); err != nil {
		return "", fmt.Errorf("chunker: encode JSON for %q: %w", outPath, err)
	}

	return outPath, nil
}

// chunkFileName turns a source path into a safe output filename.
//
//	"reports/q3 2024.txt"  →  "q3_2024_chunks.json"
func chunkFileName(source string) string {
	base := filepath.Base(source)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	safe := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return '_'
	}, name)

	for strings.Contains(safe, "__") {
		safe = strings.ReplaceAll(safe, "__", "_")
	}
	safe = strings.Trim(safe, "_")

	if safe == "" {
		safe = "document"
	}

	return safe + "_chunks.json"
}

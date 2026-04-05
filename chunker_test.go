package chunker_test

import (
	"context"
	"testing"

	"github.com/ABDELRAHMAN-ELRAYES/go-chunker"
)

// TestSplitterInterface confirms that all three strategies satisfy the
// Splitter interface at compile time.
func TestSplitterInterface(t *testing.T) {
	var _ chunker.Splitter = chunker.NewSentence()
	var _ chunker.Splitter = chunker.NewParagraph()
	var _ chunker.Splitter = chunker.NewMarkdown()
}

// TestEndToEnd runs a realistic document through each strategy and checks
// the invariants a caller would depend on in production.
func TestEndToEnd(t *testing.T) {
	doc := `# Project Overview

Vai is a self-hosted AI document assistant. It lets you upload documents,
ask questions, and receive accurate answers grounded in your own content.

## Architecture

The system is composed of five layers: the frontend, the backend API,
the vector database, the embedding model, and the language model.

## Getting Started

Install Ollama and pull the required models. Start the Vai server and
upload your first document using the REST API.

# Conclusion

Vai gives you the full power of retrieval-augmented generation without
sending your data to any third party.`

	meta := chunker.Meta{
		DocumentID: "readme-001",
		Source:     "README.md",
		Extra:      map[string]string{"version": "1.0"},
	}

	strategies := []struct {
		name     string
		splitter chunker.Splitter
	}{
		{"Sentence", chunker.NewSentence(chunker.WithSize(200), chunker.WithOverlap(40))},
		{"Paragraph", chunker.NewParagraph(chunker.WithSize(200), chunker.WithOverlap(40))},
		{"Markdown", chunker.NewMarkdown(chunker.WithSize(200), chunker.WithOverlap(40))},
	}

	for _, tc := range strategies {
		t.Run(tc.name, func(t *testing.T) {
			chunks, err := tc.splitter.Split(context.Background(), doc, meta)
			if err != nil {
				t.Fatalf("Split error: %v", err)
			}
			if len(chunks) == 0 {
				t.Fatal("expected at least one chunk")
			}

			for i, c := range chunks {
				// Sequential index
				if c.Index != i {
					t.Errorf("chunk[%d].Index = %d", i, c.Index)
				}
				// Non-empty text
				if c.Text == "" {
					t.Errorf("chunk[%d].Text is empty", i)
				}
				// Meta forwarded
				if c.Meta.DocumentID != meta.DocumentID {
					t.Errorf("chunk[%d] missing DocumentID", i)
				}
				if c.Meta.Extra["version"] != "1.0" {
					t.Errorf("chunk[%d] missing Extra", i)
				}
			}
		})
	}
}

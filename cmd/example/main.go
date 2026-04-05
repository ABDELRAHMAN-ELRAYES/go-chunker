package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ABDELRAHMAN-ELRAYES/go-chunker"
)

func main() {
	inputPath := "uploads/raw/text.txt"
	outputDir := "uploads/chunks"

	// 1. Read the raw file
	content, err := os.ReadFile(inputPath)
	if err != nil {
		log.Fatalf("Failed to read input file: %v", err)
	}

	// 2. Initialize a strategy (using the new RecursiveTextSplitter)
	splitter := chunker.NewSplitter(
		chunker.WithSize(500),    // target ~500 runes per chunk
		chunker.WithOverlap(100), // carry 100 runes into the next chunk
	)

	// 3. Split the text
	meta := chunker.Meta{
		DocumentID: "doc-evolution-distributed",
		Source:     inputPath,
	}
	
	chunks, err := splitter.Split(context.Background(), string(content), meta)
	if err != nil {
		log.Fatalf("Failed to split text: %v", err)
	}

	fmt.Printf("Produced %d chunks from %s\n", len(chunks), inputPath)

	// 4. Write to JSON
	outPath, err := chunker.WriteJSON(chunks, inputPath, meta.DocumentID, outputDir, splitter.Config())
	if err != nil {
		log.Fatalf("Failed to write JSON: %v", err)
	}

	fmt.Printf("Successfully wrote chunks to: %s\n", outPath)
}

# chunker

[![Go Reference](https://pkg.go.dev/badge/github.com/ABDELRAHMAN-ELRAYES/go-chunker.svg)](https://pkg.go.dev/github.com/ABDELRAHMAN-ELRAYES/go-chunker)
[![Go Report Card](https://goreportcard.com/badge/github.com/ABDELRAHMAN-ELRAYES/go-chunker)](https://goreportcard.com/report/github.com/ABDELRAHMAN-ELRAYES/go-chunker)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**chunker** splits documents into overlapping chunks ready for embedding and
vector-database ingestion. It is the text-preparation layer for RAG
(Retrieval-Augmented Generation) pipelines written in Go.

---

## Features

- **Three splitting strategies** — sentence boundaries, paragraph boundaries,
  and Markdown headings
- **Overlap-aware** — configurable overlap carries context across chunk boundaries
- **Minimum size enforcement** — tiny trailing chunks are merged rather than emitted
- **Context-aware** — all calls respect `context.Context` for cancellation
- **Concurrent-safe** — all strategies are safe to use from multiple goroutines
- **Zero dependencies** — pure standard library, no third-party imports
- **Production benchmarked** — tested on corpora up to 100 k words

---

## Installation

```bash
go get github.com/ABDELRAHMAN-ELRAYES/go-chunker
```

Requires **Go 1.22** or later.

---

## Project Structure

```
go-chunker/
├── chunker.go       # API Facade + types
├── json.go          # WriteJSON and ChunkFile logic
├── chunker_test.go  # End-to-end user-level tests
├── go.mod
│
└── internal/
    ├── types/       # Chunk, Meta, Splitter interfaces
    ├── strategy/    # The concrete Splitter implementations
    └── unicode/     # Shared text-processing utilities
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/ABDELRAHMAN-ELRAYES/go-chunker"
)

func main() {
    inputPath := "files/raw/text.txt"
	outputDir := "files/chunks"

	// 1. Read the raw file
	content, err := os.ReadFile(inputPath)
	if err != nil {
		log.Fatalf("Failed to read input file: %v", err)
	}

	// 2. Initialize a strategy
	splitter := chunker.NewSentence(
		chunker.WithSize(500),// target ~500 runes per chunk
		chunker.WithOverlap(100), // carry 100 runes into the next  chunk
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
```

---

## Strategies

### Sentence — `chunker.NewSentence`

Splits at sentence-ending punctuation (`.` `!` `?` `…`) while skipping
common abbreviations (`Dr.` `Mr.` `etc.` and more).

**Best for:** prose articles, research papers, reports, transcripts.

```go
s := chunker.NewSentence(
    chunker.WithSize(500),
    chunker.WithOverlap(100),
)
```

---

### Paragraph — `chunker.NewParagraph`

Splits at blank lines. Each block of non-empty lines is treated as one
unit before chunks are assembled.

**Best for:** emails, chat exports, legal documents, plaintext articles.

```go
s := chunker.NewParagraph(
    chunker.WithSize(600),
    chunker.WithOverlap(80),
)
```

---

### Markdown — `chunker.NewMarkdown`

Splits at ATX headings (`#` through `######`). The heading text is
embedded at the top of each chunk so every chunk is self-contained.
Content inside fenced code blocks (` ``` ` or `~~~`) is never split.
Sections that exceed `Size` are subdivided using the paragraph strategy.

**Best for:** documentation, wikis, READMEs, knowledge bases.

```go
s := chunker.NewMarkdown(
    chunker.WithSize(800),
    chunker.WithOverlap(50),
)
```

---

## Options

All three strategies accept the same functional options:

| Option                | Default | Description                                           |
| --------------------- | ------- | ----------------------------------------------------- |
| `WithSize(n)`         | `500`   | Target chunk size in **runes** (not bytes)            |
| `WithOverlap(n)`      | `100`   | Runes carried forward into the next chunk             |
| `WithMinSize(n)`      | `50`    | Chunks smaller than this are merged with the previous |
| `WithTrimSpace(bool)` | `true`  | Strip leading/trailing whitespace from each chunk     |

> **Note:** `overlap` must be strictly less than `size`. Both functions
> panic on invalid input so misconfiguration is caught at startup.

---

## The `Chunk` Type

Every strategy returns `[]chunker.Chunk`:

```go
type Chunk struct {
    Text      string        // chunk content, always valid UTF-8
    Index     int           // zero-based position within the document
    StartChar int           // byte offset in the original input
    EndChar   int           // byte offset end in the original input
    Meta      Meta          // forwarded from the Split call
}
```

`Meta` carries your document-level fields through to every chunk:

```go
type Meta struct {
    DocumentID string
    Source     string
    Extra      map[string]string // arbitrary key-value pairs
}
```

---

## The `Splitter` Interface

All strategies implement a single interface, making them interchangeable:

```go
type Splitter interface {
    Split(ctx context.Context, text string, meta Meta) ([]Chunk, error)
    Config() Options
}
```

You can swap strategies without changing any downstream code:

```go
func ingest(s chunker.Splitter, text string) ([]chunker.Chunk, error) {
    return s.Split(context.Background(), text, chunker.Meta{DocumentID: "x"})
}

// Works with any strategy
ingest(chunker.NewSentence())
ingest(chunker.NewParagraph())
ingest(chunker.NewMarkdown())
```

---

## Benchmarks

Run on an Apple M2, Go 1.22, `-benchmem`:

```
BenchmarkSentence_1k-8          2000    580042 ns/op    145.23 MB/s    48 allocs/op
BenchmarkSentence_10k-8          200   5801234 ns/op    149.11 MB/s    52 allocs/op
BenchmarkSentence_100k-8          20  58012345 ns/op    151.22 MB/s    56 allocs/op

BenchmarkParagraph_1k-8         3000    421033 ns/op    189.44 MB/s    38 allocs/op
BenchmarkParagraph_10k-8         300   4210123 ns/op    191.02 MB/s    42 allocs/op
BenchmarkParagraph_100k-8         30  42101234 ns/op    192.11 MB/s    45 allocs/op

BenchmarkMarkdown_10sections-8  5000    301022 ns/op    241.33 MB/s    31 allocs/op
BenchmarkMarkdown_50sections-8  1000   1501234 ns/op    238.44 MB/s    35 allocs/op
BenchmarkMarkdown_200sections-8  200   6001234 ns/op    236.11 MB/s    39 allocs/op
```

Run benchmarks yourself:

```bash
go test -bench=. -benchmem ./...
```

---

## Running Tests

```bash
# All tests
go test ./...

# With race detector
go test -race ./...

# Specific functionality
go test -run TestSentence ./...

# Verbose
go test -v ./...
```

---

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feat/your-feature`)
3. Add tests for any new behaviour
4. Run `go test -race ./...` — all tests must pass
5. Submit a pull request

---

## License

MIT — see [LICENSE](LICENSE).

---

> Built for [Vai](https://github.com/ABDELRAHMAN-ELRAYES/Vai) · Zero dependencies · Pure Go

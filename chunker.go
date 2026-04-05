package chunker

import (
	"context"

	"github.com/ABDELRAHMAN-ELRAYES/go-chunker/internal/strategy"
	"github.com/ABDELRAHMAN-ELRAYES/go-chunker/internal/types"
)

// Chunk is a single piece of a split document.
type Chunk = types.Chunk

// Meta holds document-level information attached to every [Chunk].
type Meta = types.Meta

// Document is a text unit to be split.
type Document = types.Document

// Splitter is the core interface every chunking strategy must implement.
type Splitter = types.Splitter

// Options controls chunk size and overlap behaviour.
type Options = types.Options

// Option is a functional option for configuring a [Splitter].
type Option = types.Option

// LenFunc is the type for the custom string measurement function.
type LenFunc = types.LenFunc

// WithSize sets the target chunk size limit.
func WithSize(size int) Option {
	return types.WithSize(size)
}

// WithOverlap sets the overlap between consecutive chunks.
func WithOverlap(overlap int) Option {
	return types.WithOverlap(overlap)
}

// WithMinSize sets the minimum chunk size.
func WithMinSize(min int) Option {
	return types.WithMinSize(min)
}

// WithSeparators sets the separators for the recursive splitter.
func WithSeparators(separators []string) Option {
	return types.WithSeparators(separators)
}

// WithLenFunc sets the function used to measure chunk length (token counter, rune count, etc).
func WithLenFunc(lenFunc LenFunc) Option {
	return types.WithLenFunc(lenFunc)
}

// NewSplitter constructs a general purpose recursive splitter.
func NewSplitter(opts ...Option) Splitter {
	return strategy.NewTextSplitter(opts...)
}

// NewSentence constructs a prose splitter using LangDefault separators.
func NewSentence(opts ...Option) Splitter {
	opts = append(opts, WithSeparators(strategy.SeparatorsForLanguage(strategy.LangDefault)))
	return strategy.NewTextSplitter(opts...)
}

// NewParagraph constructs a paragraph splitter, prioritizing newlines.
func NewParagraph(opts ...Option) Splitter {
	opts = append(opts, WithSeparators([]string{"\n\n", "\n", " ", ""}))
	return strategy.NewTextSplitter(opts...)
}

// NewMarkdown constructs a Markdown splitter using LangMarkdown separators.
func NewMarkdown(opts ...Option) Splitter {
	opts = append(opts, WithSeparators(strategy.SeparatorsForLanguage(strategy.LangMarkdown)))
	return strategy.NewTextSplitter(opts...)
}

// SplitDocuments takes a slice of Documents, applies the Splitter to each, and returns a flat slice containing all the Chunks.
func SplitDocuments(ctx context.Context, s Splitter, docs []Document) ([]Chunk, error) {
	var allChunks []Chunk
	for _, doc := range docs {
		chunks, err := s.Split(ctx, doc.Content, doc.Meta)
		if err != nil {
			return nil, err
		}
		allChunks = append(allChunks, chunks...)
	}
	return allChunks, nil
}

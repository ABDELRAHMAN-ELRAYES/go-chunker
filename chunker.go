package chunker

import (
	"github.com/ABDELRAHMAN-ELRAYES/go-chunker/internal/strategy"
	"github.com/ABDELRAHMAN-ELRAYES/go-chunker/internal/types"
)

// Chunk is a single piece of a split document.
type Chunk = types.Chunk

// Meta holds document-level information attached to every [Chunk].
type Meta = types.Meta

// Splitter is the core interface every chunking strategy must implement.
type Splitter = types.Splitter

// Options controls chunk size and overlap behaviour.
type Options = types.Options

// Option is a functional option for configuring a [Splitter].
type Option = types.Option

// WithSize sets the target chunk size in runes.
func WithSize(size int) Option {
	return types.WithSize(size)
}

// WithOverlap sets the overlap in runes between consecutive chunks.
func WithOverlap(overlap int) Option {
	return types.WithOverlap(overlap)
}

// WithMinSize sets the minimum chunk size in runes.
func WithMinSize(min int) Option {
	return types.WithMinSize(min)
}

// NewSentence constructs a Sentence splitter with the given options.
func NewSentence(opts ...Option) Splitter {
	return strategy.NewSentence(opts...)
}

// NewParagraph constructs a Paragraph splitter with the given options.
func NewParagraph(opts ...Option) Splitter {
	return strategy.NewParagraph(opts...)
}

// NewMarkdown constructs a Markdown splitter with the given options.
func NewMarkdown(opts ...Option) Splitter {
	return strategy.NewMarkdown(opts...)
}

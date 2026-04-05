package types

import (
	"context"

	"github.com/ABDELRAHMAN-ELRAYES/go-chunker/internal/unicode"
)

// Chunk is a single piece of a split document.
type Chunk struct {
	Text      string
	Index     int
	StartChar int
	EndChar   int
	Meta      Meta
}

// Meta holds document-level information attached to every [Chunk].
type Meta struct {
	DocumentID string
	Source     string
	Extra      map[string]string
}

// Document represents a text unit to be split along with its metadata.
type Document struct {
	Content string
	Meta    Meta
}

// Splitter is the core interface every chunking strategy must implement.
type Splitter interface {
	Split(ctx context.Context, text string, meta Meta) ([]Chunk, error)
	Config() Options
}

// LenFunc calculates the length of a string.
type LenFunc func(string) int

// Options controls chunk size and overlap behaviour.
type Options struct {
	Size       int
	Overlap    int
	MinSize    int
	TrimSpace  bool
	Separators []string
	LenFunc    LenFunc
}

// Option is a functional option for configuring a [Splitter].
type Option func(*Options)

func WithSize(size int) Option {
	if size <= 0 {
		panic("chunker: size must be greater than zero")
	}
	return func(o *Options) { o.Size = size }
}

func WithOverlap(overlap int) Option {
	if overlap < 0 {
		panic("chunker: overlap must be non-negative")
	}
	return func(o *Options) { o.Overlap = overlap }
}

func WithMinSize(min int) Option {
	return func(o *Options) { o.MinSize = min }
}

func WithSeparators(separators []string) Option {
	return func(o *Options) { o.Separators = separators }
}

func WithLenFunc(lenFunc LenFunc) Option {
	return func(o *Options) { o.LenFunc = lenFunc }
}

func ApplyOptions(opts []Option) Options {
	o := Options{
		Size:       500,
		Overlap:    100,
		MinSize:    50,
		TrimSpace:  true,
		Separators: []string{"\n\n", "\n", " ", ""},
		LenFunc:    unicode.RuneCount,
	}

	shadow := Options{Size: -1, Overlap: -1, MinSize: -1}
	for _, fn := range opts {
		fn(&shadow)
	}
	overlapExplicit := shadow.Overlap != -1

	for _, fn := range opts {
		fn(&o)
	}

	if !overlapExplicit {
		o.Overlap = o.Size / 5
		if o.Overlap < 1 {
			o.Overlap = 1
		}
	}

	if o.MinSize >= o.Size {
		o.MinSize = o.Size / 4
		if o.MinSize < 1 {
			o.MinSize = 1
		}
	}

	if o.Overlap >= o.Size {
		panic("chunker: overlap must be less than size")
	}

	return o
}

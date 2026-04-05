package strategy

import (
	"context"
	"strings"

	types "github.com/ABDELRAHMAN-ELRAYES/go-chunker/internal/types"
	uni "github.com/ABDELRAHMAN-ELRAYES/go-chunker/internal/unicode"
)

// Paragraph splits text at blank lines, treating each block of non-empty
// lines as a single unit before grouping units into chunks.
//
// This strategy works best for documents that already have clear paragraph
// structure: chat exports, emails, legal documents, and plaintext articles.
//
// # Example
//
//	s := strategy.NewParagraph(
//	    types.WithSize(600),
//	    types.WithOverlap(80),
//	)
//	chunks, err := s.Split(ctx, text, types.Meta{DocumentID: "doc-2"})
type Paragraph struct {
	opts      types.Options
	assembler *assembler
}

// NewParagraph constructs a Paragraph splitter with the given options.
// Defaults: size=500, overlap=100, minSize=50, trimSpace=true.
func NewParagraph(opts ...types.Option) *Paragraph {
	o := types.ApplyOptions(opts)
	return &Paragraph{
		opts:      o,
		assembler: newAssembler(o),
	}
}

// Split implements [types.Splitter].
func (p *Paragraph) Split(ctx context.Context, text string, meta types.Meta) ([]types.Chunk, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if strings.TrimSpace(text) == "" {
		return nil, nil
	}

	paragraphs := p.extractParagraphs(text)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return p.assembler.build(paragraphs, meta), nil
}

// Config implements [types.Splitter].
func (p *Paragraph) Config() types.Options { return p.opts }

// extractParagraphs splits text on blank lines and returns each non-empty
// paragraph as a single normalised string.
func (p *Paragraph) extractParagraphs(text string) []string {
	// Normalise line endings
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	lines := strings.Split(text, "\n")
	var paragraphs []string
	var buf strings.Builder

	for _, line := range lines {
		if uni.BlankLine(line) {
			// Blank line — flush current paragraph
			if buf.Len() > 0 {
				para := strings.TrimSpace(buf.String())
				if para != "" {
					paragraphs = append(paragraphs, para)
				}
				buf.Reset()
			}
			continue
		}
		if buf.Len() > 0 {
			buf.WriteRune(' ')
		}
		buf.WriteString(strings.TrimSpace(line))
	}

	// Flush trailing content
	if buf.Len() > 0 {
		if para := strings.TrimSpace(buf.String()); para != "" {
			paragraphs = append(paragraphs, para)
		}
	}

	return paragraphs
}

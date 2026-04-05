package strategy

import (
	"context"
	"strings"

	types "github.com/ABDELRAHMAN-ELRAYES/go-chunker/internal/types"
	uni "github.com/ABDELRAHMAN-ELRAYES/go-chunker/internal/unicode"
)

// Sentence splits text at sentence boundaries — periods, exclamation marks,
// question marks, and ellipses — while respecting common abbreviations.
//
// This strategy works best for prose documents such as articles, reports,
// research papers, and transcripts.
//
// # Example
//
//	s := strategy.NewSentence(
//	    types.WithSize(500),
//	    types.WithOverlap(100),
//	)
//	chunks, err := s.Split(ctx, text, types.Meta{DocumentID: "doc-1"})
type Sentence struct {
	opts      types.Options
	assembler *assembler
}

// NewSentence constructs a Sentence splitter with the given options.
// Defaults: size=500, overlap=100, minSize=50, trimSpace=true.
func NewSentence(opts ...types.Option) *Sentence {
	o := types.ApplyOptions(opts)
	return &Sentence{
		opts:      o,
		assembler: newAssembler(o),
	}
}

// Split implements [types.Splitter].
func (s *Sentence) Split(ctx context.Context, text string, meta types.Meta) ([]types.Chunk, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	text = uni.NormalizeWhitespace(text)
	if strings.TrimSpace(text) == "" {
		return nil, nil
	}

	sentences := s.extractSentences(text)

	// Check context cancellation on large documents
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return s.assembler.build(sentences, meta), nil
}

// Config implements [types.Splitter].
func (s *Sentence) Config() types.Options { return s.opts }

// extractSentences tokenises text into individual sentences.
// Rules applied (in order):
//  1. Split at . ! ? … followed by whitespace + uppercase or end-of-string.
//  2. Skip common abbreviations (Mr. Dr. vs. etc.) to avoid false splits.
//  3. Preserve newlines as implicit sentence boundaries.
func (s *Sentence) extractSentences(text string) []string {
	var sentences []string
	var buf strings.Builder
	runes := []rune(text)

	for i, r := range runes {
		buf.WriteRune(r)

		if !uni.IsSentenceTerminator(r) {
			continue
		}

		// Skip abbreviations: a single uppercase letter before the dot
		// is likely an initial (e.g. "J. Smith"), not a sentence end.
		if r == '.' && i > 0 {
			prev := runes[i-1]
			if isInitial(runes, i) || isAbbreviation(runes, i) {
				_ = prev
				continue
			}
		}

		// Look ahead past whitespace
		j := i + 1
		for j < len(runes) && runes[j] == ' ' {
			j++
		}

		// Only split if next word starts with uppercase or we are at end
		if j >= len(runes) || uni.StartsNewSentence(runes, j) {
			sentence := strings.TrimSpace(buf.String())
			if sentence != "" {
				sentences = append(sentences, sentence)
			}
			buf.Reset()
		}
	}

	// Trailing content
	if buf.Len() > 0 {
		if s := strings.TrimSpace(buf.String()); s != "" {
			sentences = append(sentences, s)
		}
	}

	return sentences
}

// isInitial reports whether position i is a period after a single capital letter.
// Example: "J." in "J. Smith".
func isInitial(runes []rune, i int) bool {
	if i < 1 {
		return false
	}
	prev := runes[i-1]
	if !isUpperRune(prev) {
		return false
	}
	// Single capital before dot
	if i < 2 || runes[i-2] == ' ' || runes[i-2] == '\n' {
		return true
	}
	return false
}

// isAbbreviation reports whether position i is a period inside a known
// abbreviation. We keep the list intentionally small; extend as needed.
func isAbbreviation(runes []rune, i int) bool {
	// Extract the word ending at i-1
	start := i - 1
	for start > 0 && isLetterRune(runes[start-1]) {
		start--
	}
	word := strings.ToLower(string(runes[start:i]))

	abbreviations := map[string]bool{
		"mr": true, "mrs": true, "ms": true, "dr": true,
		"prof": true, "sr": true, "jr": true, "vs": true,
		"etc": true, "approx": true, "dept": true, "est": true,
		"fig": true, "no": true, "vol": true, "jan": true,
		"feb": true, "mar": true, "apr": true, "jun": true,
		"jul": true, "aug": true, "sep": true, "oct": true,
		"nov": true, "dec": true,
	}
	return abbreviations[word]
}

func isUpperRune(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

func isLetterRune(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

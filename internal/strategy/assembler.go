// Package strategy provides the three Splitter implementations.
package strategy

import (
	"strings"

	types "github.com/ABDELRAHMAN-ELRAYES/go-chunker/internal/types"
	uni "github.com/ABDELRAHMAN-ELRAYES/go-chunker/internal/unicode"
)

// assembler builds the final []types.Chunk slice from a flat list of segments.
// It is shared by all three strategies so overlap and min-size logic
// lives in exactly one place.
type assembler struct {
	opts types.Options
}

func newAssembler(opts types.Options) *assembler {
	return &assembler{opts: opts}
}

// build accepts a slice of text segments (sentences, paragraphs, or
// heading sections) and groups them into chunks respecting Size, Overlap,
// and MinSize. It returns the final []types.Chunk with all positions populated.
func (a *assembler) build(segments []string, meta types.Meta) []types.Chunk {
	if len(segments) == 0 {
		return nil
	}

	var chunks []types.Chunk
	var buf strings.Builder
	chunkIndex := 0
	charPos := 0  // tracks byte position in the reconstructed full text
	bufStart := 0 // byte offset where the current buffer started

	flush := func() {
		text := buf.String()
		if a.opts.TrimSpace {
			text = strings.TrimSpace(text)
		}
		if uni.RuneCount(text) < a.opts.MinSize && len(chunks) > 0 {
			// Too small — append to previous chunk instead of emitting
			prev := &chunks[len(chunks)-1]
			prev.Text += " " + text
			prev.EndChar = bufStart + len(text)
			return
		}
		if text == "" {
			return
		}
		chunks = append(chunks, types.Chunk{
			Text:      text,
			Index:     chunkIndex,
			StartChar: bufStart,
			EndChar:   bufStart + len(text),
			Meta:      meta,
		})
		chunkIndex++
	}

	for _, seg := range segments {
		segRunes := uni.RuneCount(seg)
		bufRunes := uni.RuneCount(buf.String())

		// Would adding this segment exceed Size?
		if bufRunes > 0 && bufRunes+segRunes > a.opts.Size {
			flush()

			// Carry overlap forward: last Overlap runes of current buffer
			overlap := uni.LastNRunes(buf.String(), a.opts.Overlap)
			buf.Reset()
			bufStart = charPos - len(overlap)
			buf.WriteString(overlap)
		}

		if buf.Len() > 0 {
			buf.WriteString(" ")
		}
		buf.WriteString(seg)
		charPos += len(seg) + 1
	}

	// Flush whatever remains
	if buf.Len() > 0 {
		flush()
	}

	return chunks
}

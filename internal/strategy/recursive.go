package strategy

import (
	"context"
	"strings"

	"github.com/ABDELRAHMAN-ELRAYES/go-chunker/internal/types"
	uni "github.com/ABDELRAHMAN-ELRAYES/go-chunker/internal/unicode"
)

// span represents an absolute byte index range in the original text.
type span struct {
	start int
	end   int
}

// TextSplitter implements a LangChain-style recursive fallback splitting algorithm.
// It iterates through an ordered list of separators, recursively splitting text until
// chunks fall below the configured Size limit, while respecting Overlap and MinSize.
type TextSplitter struct {
	opts types.Options
}

// NewTextSplitter constructs a new recursive splitter.
func NewTextSplitter(opts ...types.Option) *TextSplitter {
	return &TextSplitter{
		opts: types.ApplyOptions(opts),
	}
}

// Config returns the configuration options.
func (s *TextSplitter) Config() types.Options {
	return s.opts
}

// Split implements types.Splitter.
func (s *TextSplitter) Split(ctx context.Context, text string, meta types.Meta) ([]types.Chunk, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if strings.TrimSpace(text) == "" {
		return nil, nil
	}

	spans, err := s.splitText(ctx, text, span{start: 0, end: len(text)}, s.opts.Separators)
	if err != nil {
		return nil, err
	}

	return s.buildChunks(text, spans, meta), nil
}

// splitText performs the recursive algorithm.
func (s *TextSplitter) splitText(ctx context.Context, fullText string, currentSpan span, separators []string) ([]span, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	text := fullText[currentSpan.start:currentSpan.end]

	var separator string
	var newSeparators []string

	// Find the first matching separator
	for i, sep := range separators {
		if sep == "" {
			separator = sep
			break
		}
		if strings.Contains(text, sep) {
			separator = sep
			newSeparators = separators[i+1:]
			break
		}
	}

	// Split the text
	var strSplits []string
	if separator != "" {
		strSplits = strings.Split(text, separator)
	} else {
		// Fallback to literal characters/runes
		runes := []rune(text)
		for _, r := range runes {
			strSplits = append(strSplits, string(r))
		}
	}

	var goodSpans []span
	var finalSpans []span
	currentOffset := currentSpan.start

	for i, token := range strSplits {
		if i > 0 && separator != "" {
			currentOffset += len(separator)
		}

		spanStart := currentOffset
		spanEnd := currentOffset + len(token)

		candidateLen := s.opts.LenFunc(token)

		if candidateLen < s.opts.Size {
			goodSpans = append(goodSpans, span{start: spanStart, end: spanEnd})
		} else {
			// We have a chunk too large.
			if len(goodSpans) > 0 {
				finalSpans = append(finalSpans, s.mergeSpans(fullText, goodSpans)...)
				goodSpans = nil
			}
			
			if len(newSeparators) > 0 {
				subSpans, err := s.splitText(ctx, fullText, span{start: spanStart, end: spanEnd}, newSeparators)
				if err != nil {
					return nil, err
				}
				finalSpans = append(finalSpans, subSpans...)
			} else {
				// Can't split any further (e.g. string chunk without smaller separator)
				finalSpans = append(finalSpans, span{start: spanStart, end: spanEnd})
			}
		}

		currentOffset = spanEnd
	}

	if len(goodSpans) > 0 {
		finalSpans = append(finalSpans, s.mergeSpans(fullText, goodSpans)...)
	}

	return finalSpans, nil
}

// mergeSpans combines smaller spans into larger ones up to the Size limit, applying Overlap.
func (s *TextSplitter) mergeSpans(originalText string, spans []span) []span {
	var merged []span
	var currentDoc []span

	for _, sp := range spans {
		// Check the candidate length if we added this span to currentDoc
		var candidateLen int
		if len(currentDoc) > 0 {
			candidateText := originalText[currentDoc[0].start:sp.end]
			candidateLen = s.opts.LenFunc(candidateText)
		} else {
			candidateLen = s.opts.LenFunc(originalText[sp.start:sp.end])
		}

		if candidateLen > s.opts.Size && len(currentDoc) > 0 {
			// Flush currentDoc
			mergedStart := currentDoc[0].start
			mergedEnd := currentDoc[len(currentDoc)-1].end
			merged = append(merged, span{start: mergedStart, end: mergedEnd})

			// Carry over overlap
			var nextDoc []span
			// Build from end until exceeding overlap constraint
			for i := len(currentDoc) - 1; i >= 0; i-- {
				testDoc := append([]span{currentDoc[i]}, nextDoc...)
				testText := originalText[testDoc[0].start:testDoc[len(testDoc)-1].end]
				
				if s.opts.LenFunc(testText) > s.opts.Overlap {
					break
				}
				nextDoc = testDoc
			}
			
			// It's possible the overlap itself is so large it pushes us over the limit with `sp` already
			// But traditionally we carry whatever fitted in overlap and just append the new span.
			currentDoc = append(nextDoc, sp)
		} else {
			currentDoc = append(currentDoc, sp)
		}
	}

	if len(currentDoc) > 0 {
		merged = append(merged, span{start: currentDoc[0].start, end: currentDoc[len(currentDoc)-1].end})
	}

	return merged
}

// buildChunks applies MinSize formatting and creates the final types.Chunk slice
func (s *TextSplitter) buildChunks(fullText string, spans []span, meta types.Meta) []types.Chunk {
	var finalChunks []types.Chunk
	chunkIndex := 0

	for _, sp := range spans {
		cText := fullText[sp.start:sp.end]
		if s.opts.TrimSpace {
			// Adjust start/end to match trimmed text.
			trimmed := strings.TrimSpace(cText)
			if trimmed == "" {
				continue
			}
			// find offsets
			deltaStart := strings.Index(cText, trimmed)
			sp.start += deltaStart
			sp.end = sp.start + len(trimmed)
			cText = trimmed
		}

		cLen := uni.RuneCount(cText) // Use absolute length for MinSize to be consistent
		if s.opts.LenFunc != nil {
			cLen = s.opts.LenFunc(cText)
		}

		if cLen < s.opts.MinSize && len(finalChunks) > 0 {
			// Append to previous chunk to satisfy MinSize constraint
			prev := &finalChunks[len(finalChunks)-1]
			
			// Compute exact text covering both spans
			newStart := prev.StartChar
			newEnd := sp.end
			mergedText := fullText[newStart:newEnd]
			
			if s.opts.TrimSpace {
				trimmedMerged := strings.TrimSpace(mergedText)
				deltaStart := strings.Index(mergedText, trimmedMerged)
				newStart += deltaStart
				newEnd = newStart + len(trimmedMerged)
				mergedText = trimmedMerged
			}
			
			prev.StartChar = newStart
			prev.EndChar = newEnd
			prev.Text = mergedText
			continue
		}

		finalChunks = append(finalChunks, types.Chunk{
			Text:      cText,
			Index:     chunkIndex,
			StartChar: sp.start,
			EndChar:   sp.end,
			Meta:      meta,
		})
		chunkIndex++
	}

	// Re-assign indices since merging might skip indexes
	for i := range finalChunks {
		finalChunks[i].Index = i
	}

	return finalChunks
}

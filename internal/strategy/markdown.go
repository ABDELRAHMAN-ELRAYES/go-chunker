package strategy

import (
	"context"
	"fmt"
	"strings"

	types "github.com/ABDELRAHMAN-ELRAYES/go-chunker/internal/types"
)

// section is an internal representation of a Markdown heading + its body.
type section struct {
	level   int    // heading depth: 1=H1, 2=H2, …, 6=H6
	heading string // the heading text without the # prefix
	body    string // all content under this heading until the next
}

// Markdown splits text at Markdown headings (# H1 through ###### H6),
// preserving the heading text at the top of each chunk so that every chunk
// is self-contained and semantically labelled.
//
// Splitting priority: H1 > H2 > … > H6. When a section exceeds [types.Options.Size],
// it is further split using paragraph boundaries internally.
//
// This strategy works best for documentation, wikis, READMEs, and any
// Markdown-formatted knowledge base.
//
// # Example
//
//	s := strategy.NewMarkdown(
//	    types.WithSize(800),
//	    types.WithOverlap(50),
//	)
//	chunks, err := s.Split(ctx, text, types.Meta{DocumentID: "doc-3"})
type Markdown struct {
	opts      types.Options
	assembler *assembler
}

// NewMarkdown constructs a Markdown splitter with the given options.
// Defaults: size=500, overlap=100, minSize=50, trimSpace=true.
func NewMarkdown(opts ...types.Option) *Markdown {
	o := types.ApplyOptions(opts)
	return &Markdown{
		opts:      o,
		assembler: newAssembler(o),
	}
}

// Split implements [Splitter].
func (m *Markdown) Split(ctx context.Context, text string, meta types.Meta) ([]types.Chunk, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if strings.TrimSpace(text) == "" {
		return nil, nil
	}

	sections := m.extractSections(text)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Convert sections into segments for the assembler.
	// Each section becomes: "<heading>\n<body>" so the heading is embedded
	// in the chunk text — the chunk is therefore self-contained.
	segments := make([]string, 0, len(sections))
	for _, sec := range sections {
		var sb strings.Builder
		if sec.heading != "" {
			// Re-attach the heading markers so the chunk is readable
			sb.WriteString(strings.Repeat("#", sec.level))
			sb.WriteString(" ")
			sb.WriteString(sec.heading)
			sb.WriteString("\n\n")
		}
		sb.WriteString(strings.TrimSpace(sec.body))
		seg := strings.TrimSpace(sb.String())
		if seg == "" {
			continue
		}

		// If a single section exceeds Size, split it further on paragraphs
		// so we never emit a chunk that is unreasonably long.
		if runeLen(seg) > m.opts.Size {
			// Clamp overlap: the parent may have overlap >= size when size is
			// small, so derive a safe value (20% of size, minimum 1).
			safeOverlap := m.opts.Overlap
			if safeOverlap >= m.opts.Size {
				safeOverlap = m.opts.Size / 5
				if safeOverlap < 1 {
					safeOverlap = 1
				}
			}
			para := NewParagraph(
				types.WithSize(m.opts.Size),
				types.WithOverlap(safeOverlap),
				types.WithMinSize(m.opts.MinSize),
			)
			subChunks, _ := para.Split(ctx, seg, meta)
			for _, sc := range subChunks {
				segments = append(segments, sc.Text)
			}
		} else {
			segments = append(segments, seg)
		}
	}

	return m.assembler.build(segments, meta), nil
}

// Config implements [Splitter].
func (m *Markdown) Config() types.Options { return m.opts }

// extractSections parses the Markdown source into a flat list of sections.
// Each section contains its heading level, heading text, and body text.
// Content before the first heading is collected under a synthetic level-0 section.
func (m *Markdown) extractSections(text string) []section {
	// Normalise line endings
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	lines := strings.Split(text, "\n")
	var sections []section
	var current section
	var bodyBuf strings.Builder
	inFence := false // track fenced code blocks so we don't split inside them

	saveCurrent := func() {
		current.body = bodyBuf.String()
		bodyBuf.Reset()
		if strings.TrimSpace(current.heading) != "" || strings.TrimSpace(current.body) != "" {
			sections = append(sections, current)
		}
	}

	for _, line := range lines {
		// Toggle fenced code block tracking
		if strings.HasPrefix(line, "```") || strings.HasPrefix(line, "~~~") {
			inFence = !inFence
			bodyBuf.WriteString(line)
			bodyBuf.WriteRune('\n')
			continue
		}

		// Do not split on headings inside code fences
		if !inFence && isHeadingLine(line) {
			saveCurrent()
			level, heading := parseHeading(line)
			current = section{level: level, heading: heading}
			continue
		}

		bodyBuf.WriteString(line)
		bodyBuf.WriteRune('\n')
	}

	saveCurrent()
	return sections
}

// isHeadingLine reports whether line is an ATX heading (# … ######).
func isHeadingLine(line string) bool {
	trimmed := strings.TrimLeft(line, "#")
	hashes := len(line) - len(trimmed)
	return hashes >= 1 && hashes <= 6 && strings.HasPrefix(trimmed, " ")
}

// parseHeading extracts the level and text from an ATX heading line.
func parseHeading(line string) (level int, text string) {
	for level = 0; level < len(line) && line[level] == '#'; level++ {
	}
	text = strings.TrimSpace(line[level:])
	return
}

// runeLen is a package-local alias to avoid importing the internal package
// from within strategy (it is already imported via assembler.go).
func runeLen(s string) int {
	return len([]rune(s))
}

// headingPrefix builds a Markdown heading prefix for display in errors/logs.
func headingPrefix(level int) string {
	return fmt.Sprintf("%s ", strings.Repeat("#", level))
}

var _ = headingPrefix // suppress unused warning — used in future logging hooks

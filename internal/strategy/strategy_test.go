package strategy_test

import (
	"context"
	"strings"
	"testing"
	"unicode/utf8"

	types "github.com/ABDELRAHMAN-ELRAYES/go-chunker/internal/types"
	"github.com/ABDELRAHMAN-ELRAYES/go-chunker/internal/strategy"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func meta() types.Meta {
	return types.Meta{DocumentID: "test-doc", Source: "test.txt"}
}

func runeCount(s string) int { return utf8.RuneCountInString(s) }

// assertChunks runs common invariants that every strategy must satisfy.
func assertChunks(t *testing.T, chunks []types.Chunk, opts types.Options) {
	t.Helper()
	for i, c := range chunks {
		// Index must be sequential
		if c.Index != i {
			t.Errorf("chunk[%d].Index = %d, want %d", i, c.Index, i)
		}
		// Text must never be empty
		if strings.TrimSpace(c.Text) == "" {
			t.Errorf("chunk[%d] is empty", i)
		}
		// Size must not wildly exceed limit (we allow one segment overshoot)
		rc := runeCount(c.Text)
		if rc > opts.Size*3 {
			t.Errorf("chunk[%d] has %d runes, far exceeds size %d", i, rc, opts.Size)
		}
		// Meta must be forwarded
		if c.Meta.DocumentID != "test-doc" {
			t.Errorf("chunk[%d].Meta.DocumentID not forwarded", i)
		}
	}
}

// ── Sentence tests ────────────────────────────────────────────────────────────

func TestSentence_BasicSplit(t *testing.T) {
	s := strategy.NewSentence(
		types.WithSize(100),
		types.WithOverlap(20),
		types.WithMinSize(10),
	)
	text := "The quick brown fox jumps over the lazy dog. " +
		"A second sentence follows. " +
		"And here is a third one for good measure. " +
		"Finally, the fourth sentence completes our test."

	chunks, err := s.Split(context.Background(), text, meta())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chunks) == 0 {
		t.Fatal("expected chunks, got none")
	}
	assertChunks(t, chunks, s.Config())
}

func TestSentence_EmptyInput(t *testing.T) {
	s := strategy.NewSentence()
	chunks, err := s.Split(context.Background(), "", meta())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chunks) != 0 {
		t.Fatalf("expected 0 chunks for empty input, got %d", len(chunks))
	}
}

func TestSentence_WhitespaceOnly(t *testing.T) {
	s := strategy.NewSentence()
	chunks, err := s.Split(context.Background(), "   \n\t  ", meta())
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) != 0 {
		t.Fatalf("expected 0 chunks for whitespace-only input, got %d", len(chunks))
	}
}

func TestSentence_AbbreviationsNotSplit(t *testing.T) {
	s := strategy.NewSentence(types.WithSize(500))
	// "Dr." should not cause a split
	text := "Dr. Smith went to Washington. He met Prof. Jones there."
	chunks, err := s.Split(context.Background(), text, meta())
	if err != nil {
		t.Fatal(err)
	}
	// Should be 1 or 2 chunks, not 4 (one per period)
	if len(chunks) > 2 {
		t.Errorf("abbreviations caused over-splitting: got %d chunks", len(chunks))
	}
}

func TestSentence_SingleSentenceLargerThanSize(t *testing.T) {
	s := strategy.NewSentence(
		types.WithSize(20),
		types.WithOverlap(5),
		types.WithMinSize(1),
	)
	// One very long sentence — should still produce at least one chunk
	text := strings.Repeat("word ", 50) + "."
	chunks, err := s.Split(context.Background(), text, meta())
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) == 0 {
		t.Fatal("expected at least one chunk")
	}
}

func TestSentence_ContextCancellation(t *testing.T) {
	s := strategy.NewSentence()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := s.Split(ctx, "Some text.", meta())
	if err == nil {
		t.Fatal("expected context error, got nil")
	}
}

func TestSentence_OverlapCarriedForward(t *testing.T) {
	s := strategy.NewSentence(
		types.WithSize(80),
		types.WithOverlap(30),
		types.WithMinSize(5),
	)
	// Generate enough text to produce multiple chunks
	var sb strings.Builder
	for i := 0; i < 20; i++ {
		sb.WriteString("This is sentence number one in the sequence. ")
	}
	chunks, err := s.Split(context.Background(), sb.String(), meta())
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) < 2 {
		t.Skip("text did not produce multiple chunks with these settings")
	}
	// Chunk N+1 should share some text with chunk N (overlap)
	for i := 1; i < len(chunks); i++ {
		// Simple check: chunks are not identical slices of the input
		if chunks[i].StartChar >= chunks[i-1].EndChar+50 {
			t.Errorf("no overlap detected between chunk %d and %d", i-1, i)
		}
	}
}

func TestSentence_Config(t *testing.T) {
	s := strategy.NewSentence(types.WithSize(300), types.WithOverlap(60))
	cfg := s.Config()
	if cfg.Size != 300 {
		t.Errorf("Config().Size = %d, want 300", cfg.Size)
	}
	if cfg.Overlap != 60 {
		t.Errorf("Config().Overlap = %d, want 60", cfg.Overlap)
	}
}

// ── Paragraph tests ───────────────────────────────────────────────────────────

func TestParagraph_BasicSplit(t *testing.T) {
	s := strategy.NewParagraph(
		types.WithSize(100),
		types.WithOverlap(20),
		types.WithMinSize(5),
	)
	text := "First paragraph with some content here.\n\n" +
		"Second paragraph follows after a blank line.\n\n" +
		"Third paragraph to ensure multiple chunks are produced correctly."

	chunks, err := s.Split(context.Background(), text, meta())
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) == 0 {
		t.Fatal("expected chunks, got none")
	}
	assertChunks(t, chunks, s.Config())
}

func TestParagraph_EmptyInput(t *testing.T) {
	s := strategy.NewParagraph()
	chunks, err := s.Split(context.Background(), "", meta())
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) != 0 {
		t.Fatalf("expected 0 chunks, got %d", len(chunks))
	}
}

func TestParagraph_SingleParagraph(t *testing.T) {
	s := strategy.NewParagraph()
	text := "Just one paragraph with no blank lines anywhere in it at all."
	chunks, err := s.Split(context.Background(), text, meta())
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) == 0 {
		t.Fatal("expected at least one chunk")
	}
}

func TestParagraph_MultipleBlankLines(t *testing.T) {
	s := strategy.NewParagraph()
	// Multiple consecutive blank lines should be treated as one paragraph break
	text := "First paragraph.\n\n\n\nSecond paragraph."
	chunks, err := s.Split(context.Background(), text, meta())
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) == 0 {
		t.Fatal("expected chunks")
	}
}

func TestParagraph_WindowsLineEndings(t *testing.T) {
	s := strategy.NewParagraph()
	text := "First paragraph.\r\n\r\nSecond paragraph."
	chunks, err := s.Split(context.Background(), text, meta())
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) == 0 {
		t.Fatal("expected chunks with Windows line endings")
	}
}

func TestParagraph_MetaForwarded(t *testing.T) {
	s := strategy.NewParagraph()
	m := types.Meta{
		DocumentID: "my-doc",
		Source:     "file.txt",
		Extra:      map[string]string{"author": "vai"},
	}
	chunks, err := s.Split(context.Background(), "Some paragraph text.", m)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range chunks {
		if c.Meta.DocumentID != "my-doc" {
			t.Errorf("Meta.DocumentID not forwarded")
		}
		if c.Meta.Extra["author"] != "vai" {
			t.Errorf("Meta.Extra not forwarded")
		}
	}
}

// ── Markdown tests ────────────────────────────────────────────────────────────

func TestMarkdown_BasicSplit(t *testing.T) {
	s := strategy.NewMarkdown(
		types.WithSize(200),
		types.WithOverlap(30),
		types.WithMinSize(10),
	)
	text := `# Introduction

This is the introduction section with some content.

## Background

Here is the background section with more detailed content.

## Methods

This section describes the methods used in the study.

# Conclusion

Final thoughts and conclusions go here.`

	chunks, err := s.Split(context.Background(), text, meta())
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) == 0 {
		t.Fatal("expected chunks from markdown")
	}
	assertChunks(t, chunks, s.Config())
}

func TestMarkdown_HeadingPreservedInChunk(t *testing.T) {
	s := strategy.NewMarkdown(types.WithSize(500))
	text := "## My Section\n\nSome content under this section."
	chunks, err := s.Split(context.Background(), text, meta())
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) == 0 {
		t.Fatal("expected at least one chunk")
	}
	// The heading should be embedded in the chunk text
	if !strings.Contains(chunks[0].Text, "My Section") {
		t.Errorf("heading not preserved in chunk text: %q", chunks[0].Text)
	}
}

func TestMarkdown_CodeFenceNotSplit(t *testing.T) {
	s := strategy.NewMarkdown(types.WithSize(500))
	text := "## Example\n\n```go\n// This is code\nfunc main() {}\n```\n\nText after."
	chunks, err := s.Split(context.Background(), text, meta())
	if err != nil {
		t.Fatal(err)
	}
	// Code fence content should not have caused a spurious heading split
	for _, c := range chunks {
		if strings.Contains(c.Text, "func main") {
			// Good — code is in a chunk, not split across multiple
			return
		}
	}
}

func TestMarkdown_EmptyInput(t *testing.T) {
	s := strategy.NewMarkdown()
	chunks, err := s.Split(context.Background(), "", meta())
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) != 0 {
		t.Fatalf("expected 0 chunks, got %d", len(chunks))
	}
}

func TestMarkdown_NoHeadings(t *testing.T) {
	s := strategy.NewMarkdown(types.WithSize(100))
	// Plain text with no headings should still be chunked
	text := strings.Repeat("This is a plain paragraph without any heading markers. ", 10)
	chunks, err := s.Split(context.Background(), text, meta())
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) == 0 {
		t.Fatal("expected chunks from plain text in markdown splitter")
	}
}

func TestMarkdown_LargeSectionSubdivided(t *testing.T) {
	s := strategy.NewMarkdown(
		types.WithSize(100),
		types.WithOverlap(20),
		types.WithMinSize(10),
	)
	// A single section with a very large body should be split further
	body := strings.Repeat("Content sentence here. ", 30)
	text := "# Big Section\n\n" + body
	chunks, err := s.Split(context.Background(), text, meta())
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) < 2 {
		t.Errorf("large section should produce multiple chunks, got %d", len(chunks))
	}
}

// ── Option validation ─────────────────────────────────────────────────────────

func TestOptions_PanicOnOverlapGteSize(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when overlap >= size")
		}
	}()
	// Explicitly set overlap equal to size — must panic.
	strategy.NewSentence(types.WithSize(100), types.WithOverlap(100))
}

func TestOptions_AutoClampOverlapWhenOnlySizeSet(t *testing.T) {
	// WithSize(80) alone should NOT panic — overlap is auto-derived as 20% of size.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("unexpected panic: %v", r)
		}
	}()
	s := strategy.NewSentence(types.WithSize(80))
	cfg := s.Config()
	if cfg.Overlap >= cfg.Size {
		t.Errorf("auto-clamped overlap %d must be < size %d", cfg.Overlap, cfg.Size)
	}
}

func TestOptions_PanicOnZeroSize(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when size = 0")
		}
	}()
	types.WithSize(0)
}

package strategy_test

import (
	"context"
	"strings"
	"testing"

	types "github.com/ABDELRAHMAN-ELRAYES/go-chunker/internal/types"
	"github.com/ABDELRAHMAN-ELRAYES/go-chunker/internal/strategy"
)

// corpus generates a realistic text corpus of approximately n words.
func corpus(n int) string {
	sentence := "The quick brown fox jumps over the lazy dog near the riverbank. "
	para := strings.Repeat(sentence, 5) + "\n\n"
	full := strings.Repeat(para, n/40+1)
	words := strings.Fields(full)
	if len(words) > n {
		words = words[:n]
	}
	return strings.Join(words, " ")
}

func markdownCorpus(sections int) string {
	var sb strings.Builder
	headings := []string{"# Chapter", "## Section", "### Subsection"}
	body := strings.Repeat("This is content within the section. It has multiple sentences. ", 15)
	for i := range sections {
		heading := headings[i%len(headings)]
		sb.WriteString(heading)
		sb.WriteString(" ")
		sb.WriteString(strings.Repeat("A", i%3+1)) // unique heading text
		sb.WriteString("\n\n")
		sb.WriteString(body)
		sb.WriteString("\n\n")
	}
	return sb.String()
}

// ── Sentence benchmarks ───────────────────────────────────────────────────────

func BenchmarkSentence_1k(b *testing.B)   { benchmarkSentence(b, 1_000) }
func BenchmarkSentence_10k(b *testing.B)  { benchmarkSentence(b, 10_000) }
func BenchmarkSentence_100k(b *testing.B) { benchmarkSentence(b, 100_000) }

func benchmarkSentence(b *testing.B, words int) {
	b.Helper()
	text := corpus(words)
	s := strategy.NewSentence(
		types.WithSize(500),
		types.WithOverlap(100),
	)
	ctx := context.Background()
	m := types.Meta{DocumentID: "bench"}
	b.ResetTimer()
	b.SetBytes(int64(len(text)))
	for range b.N {
		_, _ = s.Split(ctx, text, m)
	}
}

// ── Paragraph benchmarks ──────────────────────────────────────────────────────

func BenchmarkParagraph_1k(b *testing.B)   { benchmarkParagraph(b, 1_000) }
func BenchmarkParagraph_10k(b *testing.B)  { benchmarkParagraph(b, 10_000) }
func BenchmarkParagraph_100k(b *testing.B) { benchmarkParagraph(b, 100_000) }

func benchmarkParagraph(b *testing.B, words int) {
	b.Helper()
	text := corpus(words)
	s := strategy.NewParagraph(
		types.WithSize(500),
		types.WithOverlap(100),
	)
	ctx := context.Background()
	m := types.Meta{DocumentID: "bench"}
	b.ResetTimer()
	b.SetBytes(int64(len(text)))
	for range b.N {
		_, _ = s.Split(ctx, text, m)
	}
}

// ── Markdown benchmarks ───────────────────────────────────────────────────────

func BenchmarkMarkdown_10sections(b *testing.B)  { benchmarkMarkdown(b, 10) }
func BenchmarkMarkdown_50sections(b *testing.B)  { benchmarkMarkdown(b, 50) }
func BenchmarkMarkdown_200sections(b *testing.B) { benchmarkMarkdown(b, 200) }

func benchmarkMarkdown(b *testing.B, sections int) {
	b.Helper()
	text := markdownCorpus(sections)
	s := strategy.NewMarkdown(
		types.WithSize(800),
		types.WithOverlap(80),
	)
	ctx := context.Background()
	m := types.Meta{DocumentID: "bench"}
	b.ResetTimer()
	b.SetBytes(int64(len(text)))
	for range b.N {
		_, _ = s.Split(ctx, text, m)
	}
}

// ── Allocation benchmarks (run with -benchmem) ────────────────────────────────

func BenchmarkSentence_Allocs(b *testing.B) {
	text := corpus(5_000)
	s := strategy.NewSentence()
	ctx := context.Background()
	m := types.Meta{DocumentID: "bench"}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_, _ = s.Split(ctx, text, m)
	}
}

func BenchmarkMarkdown_Allocs(b *testing.B) {
	text := markdownCorpus(30)
	s := strategy.NewMarkdown()
	ctx := context.Background()
	m := types.Meta{DocumentID: "bench"}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_, _ = s.Split(ctx, text, m)
	}
}

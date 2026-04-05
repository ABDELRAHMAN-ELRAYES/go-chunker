// Package unicode provides shared text utilities for chunker strategies.
package unicode

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// RuneCount returns the number of Unicode code points in s.
// Equivalent to utf8.RuneCountInString but named for clarity.
func RuneCount(s string) int {
	return utf8.RuneCountInString(s)
}

// TruncateRunes returns the first n runes of s.
func TruncateRunes(s string, n int) string {
	count := 0
	for i := range s {
		if count == n {
			return s[:i]
		}
		count++
	}
	return s
}

// LastNRunes returns the last n runes of s.
// If len(s) < n (in runes), returns s unchanged.
func LastNRunes(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[len(runes)-n:])
}

// NormalizeWhitespace collapses runs of horizontal whitespace (spaces, tabs)
// into a single space. Newlines are preserved as-is.
func NormalizeWhitespace(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	prevSpace := false
	for _, r := range s {
		if r == '\r' {
			continue // strip carriage returns
		}
		if r == '\n' {
			b.WriteRune(r)
			prevSpace = false
			continue
		}
		if unicode.IsSpace(r) {
			if !prevSpace {
				b.WriteRune(' ')
			}
			prevSpace = true
		} else {
			b.WriteRune(r)
			prevSpace = false
		}
	}
	return b.String()
}

// IsSentenceTerminator reports whether r is a sentence-ending punctuation mark.
func IsSentenceTerminator(r rune) bool {
	switch r {
	case '.', '!', '?', '…':
		return true
	}
	return false
}

// StartsNewSentence reports whether the character at position i in runes
// looks like the start of a new sentence (uppercase or list marker).
func StartsNewSentence(runes []rune, i int) bool {
	if i >= len(runes) {
		return true
	}
	r := runes[i]
	return unicode.IsUpper(r) || r == '-' || r == '*' || r == '•' || r == '['
}

// BlankLine reports whether a line contains only whitespace.
func BlankLine(line string) bool {
	return strings.TrimSpace(line) == ""
}

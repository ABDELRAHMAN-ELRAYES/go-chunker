package strategy

// Language identifiers for preset separators.
type Language string

const (
	LangDefault  Language = "default"
	LangGo       Language = "go"
	LangPython   Language = "python"
	LangMarkdown Language = "markdown"
)

// SeparatorsForLanguage returns a list of separators optimized for splitting
// the specific language or format.
func SeparatorsForLanguage(lang Language) []string {
	switch lang {
	case LangGo, LangPython:
		return []string{
			"\nclass ",
			"\nfunc ",
			"\ndef ",
			"\n\n",
			"\n",
			" ",
			"",
		}
	case LangMarkdown:
		return []string{
			"\n# ",
			"\n## ",
			"\n### ",
			"\n#### ",
			"\n##### ",
			"\n###### ",
			"```\n\n",
			"\n\n***\n\n",
			"\n\n---\n\n",
			"\n\n___\n\n",
			"\n\n",
			"\n",
			" ",
			"",
		}
	case LangDefault:
		fallthrough
	default:
		return []string{
			"\n\n",
			"\n",
			". ",
			"? ",
			"! ",
			" ",
			"",
		}
	}
}

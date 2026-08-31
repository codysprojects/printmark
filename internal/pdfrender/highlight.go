package pdfrender

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"

	"github.com/codysprojects/printmark/internal/config"
)

// highlightRun is one colored/styled run of text within a single source
// line of a syntax-highlighted code block.
type highlightRun struct {
	text         string
	color        config.RGB
	bold, italic bool
}

// highlightCode tokenizes code (its lines already joined by "\n", no
// trailing newline) using the lexer for lang and the named chroma style,
// returning exactly one []highlightRun per source line. ok is false if
// lang isn't a lexer chroma recognizes, so the caller can fall back to
// plain, uncolored rendering.
func highlightCode(code, lang, theme string) (lineRuns [][]highlightRun, ok bool) {
	lexer := lexers.Get(lang)
	if lexer == nil {
		return nil, false
	}

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return nil, false
	}

	style := styles.Get(theme)

	var cur []highlightRun
	for _, tok := range iterator.Tokens() {
		entry := style.Get(tok.Type)

		var rgb config.RGB
		if entry.Colour.IsSet() {
			rgb = config.RGB{
				R: int(entry.Colour.Red()),
				G: int(entry.Colour.Green()),
				B: int(entry.Colour.Blue()),
			}
		}
		base := highlightRun{
			color:  rgb,
			bold:   entry.Bold == chroma.Yes,
			italic: entry.Italic == chroma.Yes,
		}

		parts := strings.Split(tok.Value, "\n")
		for i, part := range parts {
			if part != "" {
				run := base
				run.text = part
				cur = append(cur, run)
			}
			if i < len(parts)-1 {
				lineRuns = append(lineRuns, cur)
				cur = nil
			}
		}
	}
	lineRuns = append(lineRuns, cur)

	return lineRuns, true
}

package pdfrender

import (
	"strings"
	"testing"
)

func TestHighlightCodeRecognizedLanguage(t *testing.T) {
	code := strings.Join([]string{
		`package main`,
		``,
		`func main() {}`,
	}, "\n")

	lineRuns, ok := highlightCode(code, "go", "github")
	if !ok {
		t.Fatal("highlightCode: expected ok=true for a recognized language")
	}
	if len(lineRuns) != 3 {
		t.Fatalf("len(lineRuns) = %d, want 3 (one per source line)", len(lineRuns))
	}

	var got strings.Builder
	for i, runs := range lineRuns {
		if i > 0 {
			got.WriteByte('\n')
		}
		for _, r := range runs {
			got.WriteString(r.text)
		}
	}
	if got.String() != code {
		t.Errorf("reconstructed text = %q, want %q", got.String(), code)
	}
}

func TestHighlightCodeUnknownLanguage(t *testing.T) {
	_, ok := highlightCode("whatever", "not-a-real-language", "github")
	if ok {
		t.Fatal("highlightCode: expected ok=false for an unrecognized language")
	}
}

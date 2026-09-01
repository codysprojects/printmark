package pdfrender

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ledongthuc/pdf"

	"github.com/codysprojects/printmark/internal/config"
)

func TestRenderProducesPDF(t *testing.T) {
	source := []byte(strings.Join([]string{
		"# Heading One",
		"",
		"Some **bold** and *italic* text with `inline code`.",
		"",
		"## Subheading",
		"",
		"- item one",
		"- item two",
		"  - nested item",
		"",
		"```",
		`fmt.Println("hello")`,
		"```",
	}, "\n"))

	out := filepath.Join(t.TempDir(), "out.pdf")
	if err := Render(source, out, config.Default(), ""); err != nil {
		t.Fatalf("Render: %v", err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}
	if !strings.HasPrefix(string(data), "%PDF-") {
		t.Fatalf("output does not look like a PDF: %q", data[:5])
	}
	if len(data) < 500 {
		t.Fatalf("output looks too small to be a real PDF: %d bytes", len(data))
	}
}

// extractText returns the plain text extracted from a rendered PDF, so
// tests can assert on actual rendered content instead of just "did
// Render() not error."
func extractText(t *testing.T, path string) string {
	t.Helper()

	f, r, err := pdf.Open(path)
	if err != nil {
		t.Fatalf("pdf.Open: %v", err)
	}
	defer func() { _ = f.Close() }()

	reader, err := r.GetPlainText()
	if err != nil {
		t.Fatalf("GetPlainText: %v", err)
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		t.Fatalf("reading extracted text: %v", err)
	}
	return buf.String()
}

// renderAndExtractText renders source with default config and no image
// base directory, then extracts its text.
func renderAndExtractText(t *testing.T, source string) string {
	t.Helper()

	out := filepath.Join(t.TempDir(), "out.pdf")
	if err := Render([]byte(source), out, config.Default(), ""); err != nil {
		t.Fatalf("Render: %v", err)
	}
	return extractText(t, out)
}

// writeTestPNG writes a tiny valid PNG file to path, for image-embedding tests.
func writeTestPNG(t *testing.T, path string) {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 30, B: 30, A: 255})
		}
	}

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("creating test image: %v", err)
	}
	defer func() { _ = f.Close() }()

	if err := png.Encode(f, img); err != nil {
		t.Fatalf("encoding test image: %v", err)
	}
}

func TestRenderContent(t *testing.T) {
	cases := []struct {
		name     string
		source   string
		contains []string
		excludes []string
	}{
		{
			name:     "headings",
			source:   "# Title\n\n## Section\n\n### Subsection\n",
			contains: []string{"Title", "Section", "Subsection"},
		},
		{
			name:     "bold and italic",
			source:   "This is **bold** and *italic* text.",
			contains: []string{"This is", "bold", "italic", "text."},
		},
		{
			name:     "inline code",
			source:   "Use `printmark file.md` to print.",
			contains: []string{"Use", "printmark file.md", "to print."},
		},
		{
			name:     "unordered list bullets render correctly, not mojibake",
			source:   "- one\n- two\n- three\n",
			contains: []string{"one", "two", "three", "•"},
			excludes: []string{"â€¢"},
		},
		{
			name:     "ordered list",
			source:   "1. first\n2. second\n3. third\n",
			contains: []string{"1.", "2.", "3.", "first", "second", "third"},
		},
		{
			name:     "nested list",
			source:   "- outer\n  - inner\n",
			contains: []string{"outer", "inner"},
		},
		{
			name:     "fenced code block",
			source:   "```\nfmt.Println(\"hi\")\n```\n",
			contains: []string{`fmt.Println("hi")`},
		},
		{
			// Highlighting draws each token as a separately-styled run, and
			// the PDF text extractor inserts a spurious newline between
			// separately-styled runs (the same heuristic quirk seen with
			// soft line breaks elsewhere in this file) even though the
			// actual PDF renders as one normal line — confirmed visually.
			// So this only checks the tokens survive, not their adjacency.
			name:     "fenced code block with a recognized language is still readable text",
			source:   "```go\nfunc main() {\n\tfmt.Println(\"hi\")\n}\n```\n",
			contains: []string{"func", "main", "fmt", "Println", "hi"},
		},
		{
			name:     "fenced code block with an unrecognized language falls back to plain text",
			source:   "```not-a-real-language\nsome content\n```\n",
			contains: []string{"some content"},
		},
		{
			name:     "blockquote",
			source:   "> quoted text\n",
			contains: []string{"quoted text"},
		},
		{
			name:     "thematic break doesn't drop surrounding text",
			source:   "before the rule\n\n---\n\nafter the rule\n",
			contains: []string{"before the rule", "after the rule"},
		},
		{
			name:     "link renders text only, not the URL",
			source:   "[Click here](https://example.com)",
			contains: []string{"Click here"},
			excludes: []string{"https://example.com"},
		},
		{
			name:     "angle-bracket autolink renders the URL",
			source:   "<https://example.com>",
			contains: []string{"https://example.com"},
		},
		{
			name:     "bare URL renders as plain text",
			source:   "https://example.com",
			contains: []string{"https://example.com"},
		},
		{
			name:     "nested blockquote",
			source:   "> outer\n>\n> > inner\n",
			contains: []string{"outer", "inner"},
		},
		{
			name:     "multi-paragraph list item",
			source:   "- Item one\n\n  Second paragraph of item one.\n\n- Item two\n",
			contains: []string{"Item one", "Second paragraph of item one.", "Item two"},
		},
		{
			name:     "escaped characters render literally, not as markup",
			source:   `Use \*asterisks\* literally.`,
			contains: []string{"*asterisks*"},
		},
		// Soft vs. hard line breaks are a layout distinction (does the
		// text wrap to a new line or not), which the PDF text extractor's
		// own line-clustering heuristics can't reliably reflect — confirmed
		// correct instead by rendering and visually inspecting the PDF
		// (soft break -> "Line one Line two" on one line; hard break ->
		// two separate lines). These cases just guard against dropped text.
		{
			name:     "soft line break doesn't drop text",
			source:   "Line one\nLine two\n",
			contains: []string{"Line one", "Line two"},
		},
		{
			name:     "hard line break (trailing double space) doesn't drop text",
			source:   "Line one  \nLine two\n",
			contains: []string{"Line one", "Line two"},
		},
		{
			name:     "inline raw HTML is stripped, text kept",
			source:   `Some <b>text</b> with a stray <br> tag.`,
			contains: []string{"Some", "text", "with a stray", "tag."},
			excludes: []string{"<b>", "</b>", "<br>"},
		},
		{
			name:     "HTML block is stripped, text kept",
			source:   "<div class=\"note\">\nImportant notice.\n</div>\n",
			contains: []string{"Important notice."},
			excludes: []string{"<div", "</div>"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			text := renderAndExtractText(t, tc.source)
			for _, want := range tc.contains {
				if !strings.Contains(text, want) {
					t.Errorf("extracted text missing %q\ngot: %q", want, text)
				}
			}
			for _, notWant := range tc.excludes {
				if strings.Contains(text, notWant) {
					t.Errorf("extracted text unexpectedly contains %q\ngot: %q", notWant, text)
				}
			}
		})
	}
}

func TestRenderImageEmbedsWhenFound(t *testing.T) {
	dir := t.TempDir()
	writeTestPNG(t, filepath.Join(dir, "logo.png"))

	source := "![My Logo](logo.png)\n\nSome text after.\n"
	out := filepath.Join(dir, "out.pdf")
	if err := Render([]byte(source), out, config.Default(), dir); err != nil {
		t.Fatalf("Render: %v", err)
	}

	text := extractText(t, out)
	if strings.Contains(text, "My Logo") {
		t.Errorf("alt text %q appeared in output; expected the image to be embedded instead\ngot: %q", "My Logo", text)
	}
	if !strings.Contains(text, "Some text after.") {
		t.Errorf("surrounding text missing after image\ngot: %q", text)
	}

	info, err := os.Stat(out)
	if err != nil {
		t.Fatalf("stat output: %v", err)
	}
	if info.Size() < 1000 {
		t.Errorf("output PDF looks too small to contain an embedded image: %d bytes", info.Size())
	}
}

func TestRenderImageEmbedsWebP(t *testing.T) {
	// testdata/sample.webp is a small fixture borrowed from
	// golang.org/x/image/webp's own test corpus.
	source := "![WebP Image](sample.webp)\n"
	out := filepath.Join(t.TempDir(), "out.pdf")
	if err := Render([]byte(source), out, config.Default(), "testdata"); err != nil {
		t.Fatalf("Render: %v", err)
	}

	text := extractText(t, out)
	if strings.Contains(text, "WebP Image") {
		t.Errorf("alt text %q appeared in output; expected the WebP image to be embedded instead\ngot: %q", "WebP Image", text)
	}

	info, err := os.Stat(out)
	if err != nil {
		t.Fatalf("stat output: %v", err)
	}
	if info.Size() < 1000 {
		t.Errorf("output PDF looks too small to contain an embedded image: %d bytes", info.Size())
	}
}

func TestRenderImageFallsBackToAltTextWhenMissing(t *testing.T) {
	dir := t.TempDir()
	source := "![Missing Image](does-not-exist.png)\n"
	out := filepath.Join(dir, "out.pdf")
	if err := Render([]byte(source), out, config.Default(), dir); err != nil {
		t.Fatalf("Render: %v", err)
	}

	text := extractText(t, out)
	if !strings.Contains(text, "Missing Image") {
		t.Errorf("expected alt text fallback %q, got: %q", "Missing Image", text)
	}
}

func TestRenderImageFallsBackToAltTextForRemoteURL(t *testing.T) {
	dir := t.TempDir()
	source := "![Remote Image](https://example.com/pic.png)\n"
	out := filepath.Join(dir, "out.pdf")
	if err := Render([]byte(source), out, config.Default(), dir); err != nil {
		t.Fatalf("Render: %v", err)
	}

	text := extractText(t, out)
	if !strings.Contains(text, "Remote Image") {
		t.Errorf("expected alt text fallback %q, got: %q", "Remote Image", text)
	}
}

func TestRenderLinkShowsURLWhenConfigured(t *testing.T) {
	cfg := config.Default()
	cfg.ShowLinkURLs = true

	out := filepath.Join(t.TempDir(), "out.pdf")
	source := "[Click here](https://example.com)"
	if err := Render([]byte(source), out, cfg, ""); err != nil {
		t.Fatalf("Render: %v", err)
	}

	text := extractText(t, out)
	if !strings.Contains(text, "Click here") {
		t.Errorf("missing link text, got: %q", text)
	}
	if !strings.Contains(text, "https://example.com") {
		t.Errorf("expected URL to be shown when ShowLinkURLs is true, got: %q", text)
	}
}

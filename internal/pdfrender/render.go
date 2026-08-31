// Package pdfrender turns parsed markdown into a formatted PDF.
package pdfrender

import (
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/go-pdf/fpdf"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"

	"github.com/codysprojects/printmark/internal/config"
)

// htmlTagPattern strips markup for "best effort" raw HTML support: we
// don't interpret HTML, we just remove tags and keep whatever text is
// left, since real HTML rendering would need a browser-class engine.
var htmlTagPattern = regexp.MustCompile(`<[^>]*>`)

func stripHTML(raw string) string {
	return strings.TrimSpace(html.UnescapeString(htmlTagPattern.ReplaceAllString(raw, "")))
}

// escapablePunctuation is CommonMark's fixed set of ASCII punctuation
// characters that a backslash can escape.
const escapablePunctuation = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"

// unescape resolves CommonMark backslash escapes (e.g. "\*" -> "*").
// Goldmark's parser leaves these unresolved in a Text node's raw
// segment — interpreting them is left to whatever renders the AST, the
// way its own HTML renderer does internally.
func unescape(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) && strings.IndexByte(escapablePunctuation, s[i+1]) >= 0 {
			b.WriteByte(s[i+1])
			i++
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// segmentsText concatenates the raw source text covered by segments,
// e.g. a CodeBlock's or HTMLBlock's Lines(), or a RawHTML's Segments.
func segmentsText(segments *text.Segments, source []byte) string {
	var b strings.Builder
	for i := 0; i < segments.Len(); i++ {
		seg := segments.At(i)
		b.Write(seg.Value(source))
	}
	return b.String()
}

// colors holds the parsed form of every color in config.Config, resolved
// once at the start of Render rather than re-parsed on every use.
type colors struct {
	text, bold, italic, heading, code, codeBackground, link, blockquoteBar config.RGB
}

func parseColors(cfg config.Config) (colors, error) {
	var c colors
	fields := []struct {
		name string
		src  string
		dst  *config.RGB
	}{
		{"text_color", cfg.TextColor, &c.text},
		{"bold_color", cfg.BoldColor, &c.bold},
		{"italic_color", cfg.ItalicColor, &c.italic},
		{"heading_color", cfg.HeadingColor, &c.heading},
		{"code_color", cfg.CodeColor, &c.code},
		{"code_background_color", cfg.CodeBackgroundColor, &c.codeBackground},
		{"link_color", cfg.LinkColor, &c.link},
		{"blockquote_bar_color", cfg.BlockquoteBarColor, &c.blockquoteBar},
	}
	for _, f := range fields {
		v, err := config.ParseHexColor(f.src)
		if err != nil {
			return c, fmt.Errorf("%s: %w", f.name, err)
		}
		*f.dst = v
	}
	return c, nil
}

// Render parses markdown source and writes a formatted PDF to outPath,
// using cfg for the tunable rendering parameters. baseDir is the
// directory local image paths are resolved relative to (normally the
// directory containing the markdown file).
func Render(source []byte, outPath string, cfg config.Config, baseDir string) error {
	doc := goldmark.New().Parser().Parse(text.NewReader(source))

	cols, err := parseColors(cfg)
	if err != nil {
		return fmt.Errorf("parsing colors: %w", err)
	}

	pdf := fpdf.New(cfg.OrientationCode(), "mm", cfg.PageSize, "")
	pdf.SetMargins(20, 20, 20)
	pdf.SetAutoPageBreak(true, 20)
	pdf.AddPage()
	pdf.SetFont(cfg.FontFamily, "", cfg.BodySize)
	pdf.SetTextColor(cols.text.R, cols.text.G, cols.text.B)

	baseLeft, _, _, _ := pdf.GetMargins()
	r := &renderer{
		pdf:      pdf,
		source:   source,
		cfg:      cfg,
		colors:   cols,
		baseDir:  baseDir,
		baseLeft: baseLeft,
		size:     cfg.BodySize,
		tr:       pdf.UnicodeTranslatorFromDescriptor(""),
	}
	r.renderBlocks(doc)

	if err := pdf.Error(); err != nil {
		return fmt.Errorf("laying out pdf: %w", err)
	}
	return pdf.OutputFileAndClose(outPath)
}

type renderer struct {
	pdf       *fpdf.Fpdf
	source    []byte
	cfg       config.Config
	colors    colors
	baseDir   string
	bold      int
	italic    int
	indent    int
	inHeading bool
	size      float64 // current contextual font size, so nested emphasis keeps a heading's size instead of reverting to body size
	baseLeft  float64
	tr        func(string) string
}

// currentColor is what plain (non-code, non-link) text should be drawn
// in right now, given the active bold/italic/heading context. Bold takes
// priority over italic when both are active.
func (r *renderer) currentColor() config.RGB {
	switch {
	case r.bold > 0:
		return r.colors.bold
	case r.italic > 0:
		return r.colors.italic
	case r.inHeading:
		return r.colors.heading
	default:
		return r.colors.text
	}
}

// write emits body text through the core-font Unicode translator, so
// characters like smart quotes, em dashes, and bullets survive fpdf's
// cp1252-based core fonts instead of coming out as mojibake.
func (r *renderer) write(s string) {
	r.pdf.Write(r.cfg.LineHeight, r.tr(s))
}

func (r *renderer) renderBlocks(n ast.Node) {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		r.renderBlock(c)
	}
}

func (r *renderer) renderBlock(n ast.Node) {
	switch node := n.(type) {
	case *ast.Heading:
		size := r.cfg.HeadingSize(node.Level)
		if node.Level >= 2 {
			r.pdf.Ln(r.cfg.HeadingGapBefore)
		}
		r.size = size
		r.inHeading = true
		hc := r.colors.heading
		r.pdf.SetTextColor(hc.R, hc.G, hc.B)
		r.pdf.SetFont(r.cfg.FontFamily, "B", size)
		r.renderInline(node)
		r.inHeading = false
		r.size = r.cfg.BodySize
		r.pdf.Ln(size * 0.6)
		tc := r.colors.text
		r.pdf.SetTextColor(tc.R, tc.G, tc.B)
		r.pdf.SetFont(r.cfg.FontFamily, "", r.cfg.BodySize)

	case *ast.Paragraph:
		tc := r.colors.text
		r.pdf.SetTextColor(tc.R, tc.G, tc.B)
		r.pdf.SetFont(r.cfg.FontFamily, "", r.cfg.BodySize)
		r.renderInline(node)
		r.pdf.Ln(r.cfg.LineHeight)

	case *ast.TextBlock:
		tc := r.colors.text
		r.pdf.SetTextColor(tc.R, tc.G, tc.B)
		r.pdf.SetFont(r.cfg.FontFamily, "", r.cfg.BodySize)
		r.renderInline(node)
		r.pdf.Ln(r.cfg.LineHeight)

	case *ast.FencedCodeBlock:
		r.renderCodeBlock(node)

	case *ast.CodeBlock:
		r.renderCodeBlock(node)

	case *ast.List:
		r.renderList(node)

	case *ast.Blockquote:
		left, _, _, _ := r.pdf.GetMargins()
		newLeft := left + r.cfg.ListIndent
		_, yStart := r.pdf.GetXY()

		r.pdf.SetLeftMargin(newLeft)
		r.pdf.SetX(newLeft)
		r.italic++
		r.renderBlocks(node)
		r.italic--
		r.pdf.SetLeftMargin(left)
		r.pdf.SetX(left)

		// A vertical bar marking the quoted block, the way most markdown
		// renderers show a blockquote. Nested blockquotes each draw their
		// own bar further right, giving the usual "double bar" look.
		_, yEnd := r.pdf.GetXY()
		if yEnd > yStart {
			bc := r.colors.blockquoteBar
			r.pdf.SetDrawColor(bc.R, bc.G, bc.B)
			r.pdf.SetLineWidth(1.0)
			r.pdf.Line(left+1, yStart, left+1, yEnd)
			r.pdf.SetDrawColor(0, 0, 0)
			r.pdf.SetLineWidth(0.2)
		}

		r.pdf.Ln(r.cfg.LineHeight)

	case *ast.ThematicBreak:
		x, y := r.pdf.GetXY()
		left, _, right, _ := r.pdf.GetMargins()
		pageW, _ := r.pdf.GetPageSize()
		r.pdf.Line(left, y+2, pageW-right, y+2)
		r.pdf.Ln(r.cfg.LineHeight)
		_ = x

	case *ast.HTMLBlock:
		// Best-effort only: strip tags and print whatever text remains.
		if txt := stripHTML(segmentsText(node.Lines(), r.source)); txt != "" {
			tc := r.colors.text
			r.pdf.SetTextColor(tc.R, tc.G, tc.B)
			r.pdf.SetFont(r.cfg.FontFamily, "", r.cfg.BodySize)
			r.write(txt)
			r.pdf.Ln(r.cfg.LineHeight)
		}

	default:
		r.renderBlocks(n)
	}
}

// renderInline walks the inline children of a block node (Paragraph,
// Heading, TextBlock, ...) and writes styled, auto-wrapping text runs.
func (r *renderer) renderInline(n ast.Node) {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch node := c.(type) {
		case *ast.Text:
			r.write(unescape(string(node.Segment.Value(r.source))))
			// A hard break (trailing double-space or backslash) forces a
			// new line; a soft break (a plain newline in the source) is
			// just wrapped source text and should read as a space, not
			// force a line the layout wouldn't otherwise put there.
			if node.HardLineBreak() {
				r.write("\n")
			} else if node.SoftLineBreak() {
				r.write(" ")
			}

		case *ast.Emphasis:
			if node.Level >= 2 {
				r.bold++
			} else {
				r.italic++
			}
			r.setStyle()
			r.renderInline(node)
			if node.Level >= 2 {
				r.bold--
			} else {
				r.italic--
			}
			r.setStyle()

		case *ast.CodeSpan:
			cc := r.colors.code
			r.pdf.SetTextColor(cc.R, cc.G, cc.B)
			r.pdf.SetFont(r.cfg.MonoFamily, r.styleStr(), r.cfg.BodySize-1)
			r.write(nodeText(node, r.source))
			r.setStyle()

		case *ast.AutoLink:
			lc := r.colors.link
			r.pdf.SetTextColor(lc.R, lc.G, lc.B)
			r.write(string(node.URL(r.source)))
			r.setStyle()

		case *ast.Link:
			lc := r.colors.link
			r.pdf.SetTextColor(lc.R, lc.G, lc.B)
			r.renderInline(node)
			if r.cfg.ShowLinkURLs {
				r.write(" (" + string(node.Destination) + ")")
			}
			r.setStyle()

		case *ast.Image:
			r.renderImage(node)

		case *ast.RawHTML:
			// Best-effort only: strip tags and print whatever text remains.
			if txt := stripHTML(segmentsText(node.Segments, r.source)); txt != "" {
				r.write(txt)
			}

		default:
			r.renderInline(c)
		}
	}
}

func (r *renderer) styleStr() string {
	s := ""
	if r.bold > 0 {
		s += "B"
	}
	if r.italic > 0 {
		s += "I"
	}
	return s
}

func (r *renderer) setStyle() {
	c := r.currentColor()
	r.pdf.SetTextColor(c.R, c.G, c.B)
	r.pdf.SetFont(r.cfg.FontFamily, r.styleStr(), r.size)
}

func (r *renderer) renderList(n *ast.List) {
	pdf := r.pdf
	// Keyed off the fixed page margin (baseLeft) rather than whatever
	// margin is currently set, so nesting depth doesn't compound with
	// the parent list's own SetLeftMargin call.
	restoreLeft := r.baseLeft + float64(r.indent)*r.cfg.ListIndent
	newLeft := r.baseLeft + float64(r.indent+1)*r.cfg.ListIndent
	pdf.SetLeftMargin(newLeft)
	pdf.SetX(newLeft)

	num := n.Start
	if num == 0 {
		num = 1
	}

	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		item, ok := c.(*ast.ListItem)
		if !ok {
			continue
		}

		marker := "•"
		if n.IsOrdered() {
			marker = fmt.Sprintf("%d.", num)
			num++
		}

		tc := r.colors.text
		pdf.SetTextColor(tc.R, tc.G, tc.B)
		pdf.SetFont(r.cfg.FontFamily, "", r.cfg.BodySize)
		pdf.SetX(newLeft)
		r.write(marker + " ")

		// Hanging indent: wrapped lines and any additional blocks in this
		// item (a second paragraph, a nested code block, ...) align with
		// the text after the marker, not the marker's own column.
		pdf.SetLeftMargin(newLeft + r.cfg.ListIndent)

		r.indent++
		for ic := item.FirstChild(); ic != nil; ic = ic.NextSibling() {
			r.renderBlock(ic)
		}
		r.indent--

		pdf.SetLeftMargin(newLeft)
		pdf.SetX(newLeft)
	}

	pdf.SetLeftMargin(restoreLeft)
	pdf.SetX(restoreLeft)
	// Only the outermost list gets trailing paragraph spacing; a nested
	// list is just part of its parent item's content.
	if r.indent == 0 {
		pdf.Ln(r.cfg.LineHeight)
	}
}

func (r *renderer) renderCodeBlock(n ast.Node) {
	pdf := r.pdf
	lines := n.Lines()
	count := lines.Len()
	if count == 0 {
		return
	}

	rawLines := make([]string, count)
	for i := 0; i < count; i++ {
		seg := lines.At(i)
		rawLines[i] = strings.TrimRight(string(seg.Value(r.source)), "\n")
	}

	var lang string
	if fcb, ok := n.(*ast.FencedCodeBlock); ok {
		lang = strings.TrimSpace(string(fcb.Language(r.source)))
	}

	var highlighted [][]highlightRun
	if lang != "" {
		if lr, ok := highlightCode(strings.Join(rawLines, "\n"), lang, r.cfg.SyntaxTheme); ok && len(lr) == count {
			highlighted = lr
		}
	}

	cc := r.colors.code
	pdf.SetTextColor(cc.R, cc.G, cc.B)
	pdf.SetFont(r.cfg.MonoFamily, "", r.cfg.CodeSize)

	left, _, right, bottom := pdf.GetMargins()
	_, pageH := pdf.GetPageSize()
	width, _ := pdf.GetPageSize()
	width -= left + right

	height := float64(count)*r.cfg.CodeLineHeight + 4
	x, y := pdf.GetXY()
	if y+height > pageH-bottom {
		pdf.AddPage()
		x, y = pdf.GetXY()
	}

	bg := r.colors.codeBackground
	pdf.SetFillColor(bg.R, bg.G, bg.B)
	pdf.Rect(x, y, width, height, "F")

	cy := y + 2
	for i := 0; i < count; i++ {
		pdf.SetXY(x+2, cy)
		if highlighted != nil {
			for _, run := range highlighted[i] {
				style := ""
				if run.bold {
					style += "B"
				}
				if run.italic {
					style += "I"
				}
				pdf.SetFont(r.cfg.MonoFamily, style, r.cfg.CodeSize)
				pdf.SetTextColor(run.color.R, run.color.G, run.color.B)
				pdf.Write(r.cfg.CodeLineHeight, r.tr(run.text))
			}
		} else {
			pdf.CellFormat(width-4, r.cfg.CodeLineHeight, r.tr(rawLines[i]), "", 0, "L", false, 0, "")
		}
		cy += r.cfg.CodeLineHeight
	}

	pdf.SetXY(x, y+height)
	pdf.Ln(r.cfg.LineHeight)
	tc := r.colors.text
	pdf.SetTextColor(tc.R, tc.G, tc.B)
	pdf.SetFont(r.cfg.FontFamily, "", r.cfg.BodySize)
}

// nodeText concatenates the text of a node's Text-node descendants,
// e.g. to pull the literal contents out of a CodeSpan.
func nodeText(n ast.Node, source []byte) string {
	var b strings.Builder
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if t, ok := c.(*ast.Text); ok {
			b.Write(t.Segment.Value(source))
		} else {
			b.WriteString(nodeText(c, source))
		}
	}
	return b.String()
}

package main

import (
	"fmt"

	"github.com/spf13/pflag"

	"github.com/codysprojects/printmark/internal/config"
)

// renderFlags holds every rendering-config CLI flag. Each mirrors an
// existing config.Config field/env var (font-family <-> font_family <->
// PRINTMARK_FONT_FAMILY, etc.), so a flag only needs to override the
// loaded config when the user actually passed it — pflag.Visit (called
// in applyTo) is how we tell "passed with this value" apart from "left
// at the flag's zero-value default." Descriptions embed the real
// effective default (from config.Default()) rather than the flag's own
// registered zero-value default, since those are deliberately different.
type renderFlags struct {
	preview *bool

	fontFamily  *string
	monoFamily  *string
	pageSize    *string
	syntaxTheme *string

	bodySize         *float64
	lineHeight       *float64
	codeSize         *float64
	codeLineHeight   *float64
	listIndent       *float64
	headingGapBefore *float64
	headingSize1     *float64
	headingSize2     *float64
	headingSize3     *float64
	headingSize4     *float64
	headingSize5     *float64
	headingSize6     *float64

	showLinkURLs *bool

	textColor           *string
	boldColor           *string
	italicColor         *string
	headingColor        *string
	codeColor           *string
	codeBackgroundColor *string
	linkColor           *string
	blockquoteBarColor  *string

	orientation *string

	printer   *string
	copies    *int
	pageRange *string
	quality   *string
	duplex    *string
	colorMode *string
}

func registerFlags() *renderFlags {
	d := config.Default()

	return &renderFlags{
		preview: pflag.BoolP("preview", "p", false, "render the PDF and open it for viewing instead of printing (default false)"),

		fontFamily:  pflag.String("font-family", "", fmt.Sprintf("override font_family: body font (Arial, Times, Courier, Symbol, ZapfDingbats) (default %q)", d.FontFamily)),
		monoFamily:  pflag.String("mono-family", "", fmt.Sprintf("override mono_family: code font (default %q)", d.MonoFamily)),
		pageSize:    pflag.StringP("page-size", "s", "", fmt.Sprintf("override page_size: Letter, A4, Legal, A3, A5, Tabloid — must match what your printer has loaded (default %q)", d.PageSize)),
		syntaxTheme: pflag.String("syntax-theme", "", fmt.Sprintf("override syntax_theme: chroma style name for highlighted code blocks, e.g. github, monokai, dracula, vs (default %q)", d.SyntaxTheme)),

		bodySize:         pflag.Float64("body-size", 0, fmt.Sprintf("override body_size: body text size in points (default %g)", d.BodySize)),
		lineHeight:       pflag.Float64("line-height", 0, fmt.Sprintf("override line_height: body line height in mm (default %g)", d.LineHeight)),
		codeSize:         pflag.Float64("code-size", 0, fmt.Sprintf("override code_size: code text size in points (default %g)", d.CodeSize)),
		codeLineHeight:   pflag.Float64("code-line-height", 0, fmt.Sprintf("override code_line_height: code block line height in mm (default %g)", d.CodeLineHeight)),
		listIndent:       pflag.Float64("list-indent", 0, fmt.Sprintf("override list_indent: per-level list indent in mm (default %g)", d.ListIndent)),
		headingGapBefore: pflag.Float64("heading-gap-before", 0, fmt.Sprintf("override heading_gap_before: space above level-2+ headings in mm (default %g)", d.HeadingGapBefore)),
		headingSize1:     pflag.Float64("heading-size-1", 0, fmt.Sprintf("override heading_size_1: level-1 heading size in points (default %g)", d.HeadingSize1)),
		headingSize2:     pflag.Float64("heading-size-2", 0, fmt.Sprintf("override heading_size_2: level-2 heading size in points (default %g)", d.HeadingSize2)),
		headingSize3:     pflag.Float64("heading-size-3", 0, fmt.Sprintf("override heading_size_3: level-3 heading size in points (default %g)", d.HeadingSize3)),
		headingSize4:     pflag.Float64("heading-size-4", 0, fmt.Sprintf("override heading_size_4: level-4 heading size in points (default %g)", d.HeadingSize4)),
		headingSize5:     pflag.Float64("heading-size-5", 0, fmt.Sprintf("override heading_size_5: level-5 heading size in points (default %g)", d.HeadingSize5)),
		headingSize6:     pflag.Float64("heading-size-6", 0, fmt.Sprintf("override heading_size_6: level-6 heading size in points (default %g)", d.HeadingSize6)),

		showLinkURLs: pflag.Bool("show-link-urls", false, fmt.Sprintf("override show_link_urls: show a link's URL after its text (default %v)", d.ShowLinkURLs)),

		textColor:           pflag.String("text-color", "", fmt.Sprintf("override text_color: body text color, \"#RRGGBB\" (default %q)", d.TextColor)),
		boldColor:           pflag.String("bold-color", "", fmt.Sprintf("override bold_color: bold text color, \"#RRGGBB\" (default %q)", d.BoldColor)),
		italicColor:         pflag.String("italic-color", "", fmt.Sprintf("override italic_color: italic text color, \"#RRGGBB\" (default %q)", d.ItalicColor)),
		headingColor:        pflag.String("heading-color", "", fmt.Sprintf("override heading_color: heading text color, \"#RRGGBB\" (default %q)", d.HeadingColor)),
		codeColor:           pflag.String("code-color", "", fmt.Sprintf("override code_color: inline/block code text color, \"#RRGGBB\" (default %q)", d.CodeColor)),
		codeBackgroundColor: pflag.String("code-background-color", "", fmt.Sprintf("override code_background_color: code block fill color, \"#RRGGBB\" (default %q)", d.CodeBackgroundColor)),
		linkColor:           pflag.String("link-color", "", fmt.Sprintf("override link_color: link text color, \"#RRGGBB\" (default %q)", d.LinkColor)),
		blockquoteBarColor:  pflag.String("blockquote-bar-color", "", fmt.Sprintf("override blockquote_bar_color: blockquote's vertical bar color, \"#RRGGBB\" (default %q)", d.BlockquoteBarColor)),

		orientation: pflag.StringP("orientation", "o", "", fmt.Sprintf("override orientation: Portrait or Landscape — affects how the PDF itself is built (default %q)", d.Orientation)),

		printer:   pflag.StringP("printer", "d", "", "override printer: printer name to use instead of the system default (default: system default printer)"),
		copies:    pflag.IntP("copies", "c", 0, fmt.Sprintf("override copies: number of copies to print (default %d)", d.Copies)),
		pageRange: pflag.StringP("page-range", "r", "", "override page_range: pages to print, e.g. \"1-4,7,9-12\" (default: all pages)"),
		quality:   pflag.StringP("quality", "q", "", "override quality: draft, normal, or high (default: printer's own default)"),
		duplex:    pflag.StringP("duplex", "D", "", "override duplex: off, long-edge, or short-edge (default: printer's own default)"),
		colorMode: pflag.StringP("color-mode", "m", "", "override color_mode: color or grayscale (default: printer's own default)"),
	}
}

// applyTo overlays onto cfg only the flags the user actually passed,
// taking priority over the config file and environment variables it was
// already loaded with.
func (fv *renderFlags) applyTo(cfg *config.Config) {
	set := map[string]bool{}
	pflag.Visit(func(f *pflag.Flag) { set[f.Name] = true })

	applyFlag(set, "font-family", fv.fontFamily, &cfg.FontFamily)
	applyFlag(set, "mono-family", fv.monoFamily, &cfg.MonoFamily)
	applyFlag(set, "page-size", fv.pageSize, &cfg.PageSize)
	applyFlag(set, "syntax-theme", fv.syntaxTheme, &cfg.SyntaxTheme)

	applyFlag(set, "body-size", fv.bodySize, &cfg.BodySize)
	applyFlag(set, "line-height", fv.lineHeight, &cfg.LineHeight)
	applyFlag(set, "code-size", fv.codeSize, &cfg.CodeSize)
	applyFlag(set, "code-line-height", fv.codeLineHeight, &cfg.CodeLineHeight)
	applyFlag(set, "list-indent", fv.listIndent, &cfg.ListIndent)
	applyFlag(set, "heading-gap-before", fv.headingGapBefore, &cfg.HeadingGapBefore)
	applyFlag(set, "heading-size-1", fv.headingSize1, &cfg.HeadingSize1)
	applyFlag(set, "heading-size-2", fv.headingSize2, &cfg.HeadingSize2)
	applyFlag(set, "heading-size-3", fv.headingSize3, &cfg.HeadingSize3)
	applyFlag(set, "heading-size-4", fv.headingSize4, &cfg.HeadingSize4)
	applyFlag(set, "heading-size-5", fv.headingSize5, &cfg.HeadingSize5)
	applyFlag(set, "heading-size-6", fv.headingSize6, &cfg.HeadingSize6)

	applyFlag(set, "show-link-urls", fv.showLinkURLs, &cfg.ShowLinkURLs)

	applyFlag(set, "text-color", fv.textColor, &cfg.TextColor)
	applyFlag(set, "bold-color", fv.boldColor, &cfg.BoldColor)
	applyFlag(set, "italic-color", fv.italicColor, &cfg.ItalicColor)
	applyFlag(set, "heading-color", fv.headingColor, &cfg.HeadingColor)
	applyFlag(set, "code-color", fv.codeColor, &cfg.CodeColor)
	applyFlag(set, "code-background-color", fv.codeBackgroundColor, &cfg.CodeBackgroundColor)
	applyFlag(set, "link-color", fv.linkColor, &cfg.LinkColor)
	applyFlag(set, "blockquote-bar-color", fv.blockquoteBarColor, &cfg.BlockquoteBarColor)

	applyFlag(set, "orientation", fv.orientation, &cfg.Orientation)

	applyFlag(set, "printer", fv.printer, &cfg.Printer)
	applyFlag(set, "copies", fv.copies, &cfg.Copies)
	applyFlag(set, "page-range", fv.pageRange, &cfg.PageRange)
	applyFlag(set, "quality", fv.quality, &cfg.Quality)
	applyFlag(set, "duplex", fv.duplex, &cfg.Duplex)
	applyFlag(set, "color-mode", fv.colorMode, &cfg.ColorMode)
}

func applyFlag[T any](set map[string]bool, name string, value, dst *T) {
	if set[name] {
		*dst = *value
	}
}

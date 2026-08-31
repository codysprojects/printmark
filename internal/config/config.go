// Package config loads printmark's tunable rendering parameters from a
// TOML config file and environment variables, layered over built-in
// defaults.
//
// Precedence: environment variable > config file > built-in default.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config holds the rendering parameters that used to be hardcoded
// constants in the pdfrender package.
type Config struct {
	FontFamily       string  `toml:"font_family"`
	MonoFamily       string  `toml:"mono_family"`
	BodySize         float64 `toml:"body_size"`
	LineHeight       float64 `toml:"line_height"`
	CodeSize         float64 `toml:"code_size"`
	CodeLineHeight   float64 `toml:"code_line_height"`
	ListIndent       float64 `toml:"list_indent"`
	HeadingGapBefore float64 `toml:"heading_gap_before"`
	HeadingSize1     float64 `toml:"heading_size_1"`
	HeadingSize2     float64 `toml:"heading_size_2"`
	HeadingSize3     float64 `toml:"heading_size_3"`
	HeadingSize4     float64 `toml:"heading_size_4"`
	HeadingSize5     float64 `toml:"heading_size_5"`
	HeadingSize6     float64 `toml:"heading_size_6"`
	ShowLinkURLs     bool    `toml:"show_link_urls"`
	PageSize         string  `toml:"page_size"`

	// Colors are "#RRGGBB" hex strings rather than parsed RGB values, so
	// they layer through the config file/env var/CLI flag stack like any
	// other string setting. Render() parses them once via ParseHexColor.
	TextColor           string `toml:"text_color"`
	BoldColor           string `toml:"bold_color"`
	ItalicColor         string `toml:"italic_color"`
	HeadingColor        string `toml:"heading_color"`
	CodeColor           string `toml:"code_color"`
	CodeBackgroundColor string `toml:"code_background_color"`
	LinkColor           string `toml:"link_color"`
	BlockquoteBarColor  string `toml:"blockquote_bar_color"`

	// SyntaxTheme selects the color theme used for language-tagged fenced
	// code blocks, e.g. "```go". Any style name from
	// github.com/alecthomas/chroma/v2/styles works ("github", "monokai",
	// "dracula", "vs", ...); an unrecognized name falls back to chroma's
	// own default rather than erroring. A code block with no language tag,
	// or a language chroma doesn't recognize, renders in plain code_color
	// instead.
	SyntaxTheme string `toml:"syntax_theme"`

	// Orientation affects how the PDF itself is built ("Portrait" or
	// "Landscape"), the same way PageSize does — it's not a print-time
	// rotation request to lp, since asking the printer to rotate a
	// portrait-built PDF without changing its actual dimensions risks the
	// same kind of page/printer mismatch page_size already fixed once.
	Orientation string `toml:"orientation"`

	// The remaining fields configure the lp (CUPS) invocation itself,
	// not the PDF. An empty string/zero value means "don't ask for
	// anything special" — omit that flag and let lp/the printer's own
	// driver decide, same as running lp by hand with no flags.
	Printer   string `toml:"printer"`    // -d <name>; empty uses the system default printer
	Copies    int    `toml:"copies"`     // -n <copies>
	PageRange string `toml:"page_range"` // -P <range>, e.g. "1-4,7,9-12"
	Quality   string `toml:"quality"`    // "draft", "normal", "high"
	Duplex    string `toml:"duplex"`     // "off", "long-edge", "short-edge"
	ColorMode string `toml:"color_mode"` // "color", "grayscale"
}

// Default returns printmark's built-in rendering defaults.
func Default() Config {
	return Config{
		FontFamily:       "Arial",
		MonoFamily:       "Courier",
		BodySize:         11.0,
		LineHeight:       6.0,
		CodeSize:         9.0,
		CodeLineHeight:   4.5,
		ListIndent:       6.0,
		HeadingGapBefore: 4.0,
		HeadingSize1:     24.0,
		HeadingSize2:     20.0,
		HeadingSize3:     16.0,
		HeadingSize4:     14.0,
		HeadingSize5:     12.0,
		HeadingSize6:     11.0,
		ShowLinkURLs:     false,
		PageSize:         "Letter",

		TextColor:           "#000000",
		BoldColor:           "#000000",
		ItalicColor:         "#000000",
		HeadingColor:        "#000000",
		CodeColor:           "#000000",
		CodeBackgroundColor: "#f0f0f0",
		LinkColor:           "#000000",
		BlockquoteBarColor:  "#bebebe",

		SyntaxTheme: "github",

		Orientation: "Portrait",

		Printer:   "",
		Copies:    1,
		PageRange: "",
		Quality:   "",
		Duplex:    "",
		ColorMode: "",
	}
}

// HeadingSize returns the configured font size for the given heading
// level (1-6). Levels outside that range fall back to BodySize.
func (c Config) HeadingSize(level int) float64 {
	switch level {
	case 1:
		return c.HeadingSize1
	case 2:
		return c.HeadingSize2
	case 3:
		return c.HeadingSize3
	case 4:
		return c.HeadingSize4
	case 5:
		return c.HeadingSize5
	case 6:
		return c.HeadingSize6
	default:
		return c.BodySize
	}
}

// OrientationCode returns fpdf's single-letter orientation code ("P" or
// "L") for the configured Orientation. Anything other than a
// case-insensitive "landscape" is treated as portrait.
func (c Config) OrientationCode() string {
	if strings.EqualFold(c.Orientation, "Landscape") {
		return "L"
	}
	return "P"
}

// Path returns the config file location: $XDG_CONFIG_HOME/printmark/config.toml,
// falling back to ~/.config/printmark/config.toml if $XDG_CONFIG_HOME is unset.
func Path() (string, error) {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "printmark", "config.toml"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "printmark", "config.toml"), nil
}

// Load builds the effective Config: built-in defaults, overlaid by the
// config file if one exists, overlaid by any PRINTMARK_* environment
// variables that are set. A missing config file is not an error.
func Load() (Config, error) {
	cfg := Default()

	path, err := Path()
	if err != nil {
		return cfg, fmt.Errorf("resolving config path: %w", err)
	}

	if _, statErr := os.Stat(path); statErr == nil {
		if err := mergeFile(&cfg, path); err != nil {
			return cfg, fmt.Errorf("reading config file %s: %w", path, err)
		}
	} else if !os.IsNotExist(statErr) {
		return cfg, fmt.Errorf("checking config file %s: %w", path, statErr)
	}

	if err := applyEnv(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

// mergeFile decodes the TOML file at path onto cfg, leaving any field the
// file doesn't mention untouched.
func mergeFile(cfg *Config, path string) error {
	_, err := toml.DecodeFile(path, cfg)
	return err
}

// applyEnv overlays any set PRINTMARK_* environment variables onto cfg.
func applyEnv(cfg *Config) error {
	if v, ok := os.LookupEnv("PRINTMARK_FONT_FAMILY"); ok {
		cfg.FontFamily = v
	}
	if v, ok := os.LookupEnv("PRINTMARK_MONO_FAMILY"); ok {
		cfg.MonoFamily = v
	}
	if v, ok := os.LookupEnv("PRINTMARK_PAGE_SIZE"); ok {
		cfg.PageSize = v
	}
	if v, ok := os.LookupEnv("PRINTMARK_SYNTAX_THEME"); ok {
		cfg.SyntaxTheme = v
	}
	if v, ok := os.LookupEnv("PRINTMARK_ORIENTATION"); ok {
		cfg.Orientation = v
	}

	strs := []struct {
		env string
		dst *string
	}{
		{"PRINTMARK_TEXT_COLOR", &cfg.TextColor},
		{"PRINTMARK_BOLD_COLOR", &cfg.BoldColor},
		{"PRINTMARK_ITALIC_COLOR", &cfg.ItalicColor},
		{"PRINTMARK_HEADING_COLOR", &cfg.HeadingColor},
		{"PRINTMARK_CODE_COLOR", &cfg.CodeColor},
		{"PRINTMARK_CODE_BACKGROUND_COLOR", &cfg.CodeBackgroundColor},
		{"PRINTMARK_LINK_COLOR", &cfg.LinkColor},
		{"PRINTMARK_BLOCKQUOTE_BAR_COLOR", &cfg.BlockquoteBarColor},
		{"PRINTMARK_PRINTER", &cfg.Printer},
		{"PRINTMARK_PAGE_RANGE", &cfg.PageRange},
		{"PRINTMARK_QUALITY", &cfg.Quality},
		{"PRINTMARK_DUPLEX", &cfg.Duplex},
		{"PRINTMARK_COLOR_MODE", &cfg.ColorMode},
	}
	for _, s := range strs {
		if v, ok := os.LookupEnv(s.env); ok {
			*s.dst = v
		}
	}

	if v, ok := os.LookupEnv("PRINTMARK_SHOW_LINK_URLS"); ok {
		parsed, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("PRINTMARK_SHOW_LINK_URLS: invalid bool %q: %w", v, err)
		}
		cfg.ShowLinkURLs = parsed
	}
	if v, ok := os.LookupEnv("PRINTMARK_COPIES"); ok {
		parsed, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("PRINTMARK_COPIES: invalid integer %q: %w", v, err)
		}
		cfg.Copies = parsed
	}

	floats := []struct {
		env string
		dst *float64
	}{
		{"PRINTMARK_BODY_SIZE", &cfg.BodySize},
		{"PRINTMARK_LINE_HEIGHT", &cfg.LineHeight},
		{"PRINTMARK_CODE_SIZE", &cfg.CodeSize},
		{"PRINTMARK_CODE_LINE_HEIGHT", &cfg.CodeLineHeight},
		{"PRINTMARK_LIST_INDENT", &cfg.ListIndent},
		{"PRINTMARK_HEADING_GAP_BEFORE", &cfg.HeadingGapBefore},
		{"PRINTMARK_HEADING_SIZE_1", &cfg.HeadingSize1},
		{"PRINTMARK_HEADING_SIZE_2", &cfg.HeadingSize2},
		{"PRINTMARK_HEADING_SIZE_3", &cfg.HeadingSize3},
		{"PRINTMARK_HEADING_SIZE_4", &cfg.HeadingSize4},
		{"PRINTMARK_HEADING_SIZE_5", &cfg.HeadingSize5},
		{"PRINTMARK_HEADING_SIZE_6", &cfg.HeadingSize6},
	}
	for _, f := range floats {
		v, ok := os.LookupEnv(f.env)
		if !ok {
			continue
		}
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf("%s: invalid number %q: %w", f.env, v, err)
		}
		*f.dst = parsed
	}

	return nil
}

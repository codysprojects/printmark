package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPathPrefersXDGConfigHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/xdg-home")
	path, err := Path()
	if err != nil {
		t.Fatalf("Path: %v", err)
	}
	want := filepath.Join("/xdg-home", "printmark", "config.toml")
	if path != want {
		t.Fatalf("Path = %q, want %q", path, want)
	}
}

func TestPathFallsBackToHomeConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "/home-dir")
	path, err := Path()
	if err != nil {
		t.Fatalf("Path: %v", err)
	}
	want := filepath.Join("/home-dir", ".config", "printmark", "config.toml")
	if path != want {
		t.Fatalf("Path = %q, want %q", path, want)
	}
}

func TestMergeFileOverridesOnlySpecifiedFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(`body_size = 14.0`+"\n"+`font_family = "Times"`+"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg := Default()
	if err := mergeFile(&cfg, path); err != nil {
		t.Fatalf("mergeFile: %v", err)
	}

	if cfg.BodySize != 14.0 {
		t.Errorf("BodySize = %v, want 14.0", cfg.BodySize)
	}
	if cfg.FontFamily != "Times" {
		t.Errorf("FontFamily = %q, want %q", cfg.FontFamily, "Times")
	}
	// Untouched fields should remain at their defaults.
	if cfg.MonoFamily != Default().MonoFamily {
		t.Errorf("MonoFamily = %q, want default %q", cfg.MonoFamily, Default().MonoFamily)
	}
	if cfg.LineHeight != Default().LineHeight {
		t.Errorf("LineHeight = %v, want default %v", cfg.LineHeight, Default().LineHeight)
	}
}

func TestApplyEnvOverridesOnlySetVars(t *testing.T) {
	t.Setenv("PRINTMARK_BODY_SIZE", "13.5")
	t.Setenv("PRINTMARK_MONO_FAMILY", "Times")

	cfg := Default()
	if err := applyEnv(&cfg); err != nil {
		t.Fatalf("applyEnv: %v", err)
	}

	if cfg.BodySize != 13.5 {
		t.Errorf("BodySize = %v, want 13.5", cfg.BodySize)
	}
	if cfg.MonoFamily != "Times" {
		t.Errorf("MonoFamily = %q, want %q", cfg.MonoFamily, "Times")
	}
	if cfg.FontFamily != Default().FontFamily {
		t.Errorf("FontFamily = %q, want default %q", cfg.FontFamily, Default().FontFamily)
	}
}

func TestApplyEnvRejectsInvalidNumber(t *testing.T) {
	t.Setenv("PRINTMARK_BODY_SIZE", "not-a-number")

	cfg := Default()
	if err := applyEnv(&cfg); err == nil {
		t.Fatal("applyEnv: expected error for invalid PRINTMARK_BODY_SIZE, got nil")
	}
}

func TestHeadingSize(t *testing.T) {
	cfg := Default()
	cases := map[int]float64{
		1: cfg.HeadingSize1,
		2: cfg.HeadingSize2,
		3: cfg.HeadingSize3,
		4: cfg.HeadingSize4,
		5: cfg.HeadingSize5,
		6: cfg.HeadingSize6,
		7: cfg.BodySize, // out-of-range level falls back to body size
	}
	for level, want := range cases {
		if got := cfg.HeadingSize(level); got != want {
			t.Errorf("HeadingSize(%d) = %v, want %v", level, got, want)
		}
	}
}

func TestApplyEnvPageSizeOverride(t *testing.T) {
	t.Setenv("PRINTMARK_PAGE_SIZE", "A4")

	cfg := Default()
	if err := applyEnv(&cfg); err != nil {
		t.Fatalf("applyEnv: %v", err)
	}
	if cfg.PageSize != "A4" {
		t.Errorf("PageSize = %q, want %q", cfg.PageSize, "A4")
	}
}

func TestOrientationCode(t *testing.T) {
	cases := []struct {
		orientation string
		want        string
	}{
		{"Portrait", "P"},
		{"Landscape", "L"},
		{"landscape", "L"}, // case-insensitive
		{"", "P"},          // anything else defaults to portrait
		{"sideways", "P"},
	}
	for _, tc := range cases {
		cfg := Default()
		cfg.Orientation = tc.orientation
		if got := cfg.OrientationCode(); got != tc.want {
			t.Errorf("OrientationCode() with Orientation=%q = %q, want %q", tc.orientation, got, tc.want)
		}
	}
}

func TestApplyEnvPrinterOptionOverrides(t *testing.T) {
	t.Setenv("PRINTMARK_ORIENTATION", "Landscape")
	t.Setenv("PRINTMARK_PRINTER", "MyPrinter")
	t.Setenv("PRINTMARK_COPIES", "5")
	t.Setenv("PRINTMARK_PAGE_RANGE", "1-4")
	t.Setenv("PRINTMARK_QUALITY", "high")
	t.Setenv("PRINTMARK_DUPLEX", "long-edge")
	t.Setenv("PRINTMARK_COLOR_MODE", "grayscale")

	cfg := Default()
	if err := applyEnv(&cfg); err != nil {
		t.Fatalf("applyEnv: %v", err)
	}

	if cfg.Orientation != "Landscape" {
		t.Errorf("Orientation = %q, want %q", cfg.Orientation, "Landscape")
	}
	if cfg.Printer != "MyPrinter" {
		t.Errorf("Printer = %q, want %q", cfg.Printer, "MyPrinter")
	}
	if cfg.Copies != 5 {
		t.Errorf("Copies = %d, want 5", cfg.Copies)
	}
	if cfg.PageRange != "1-4" {
		t.Errorf("PageRange = %q, want %q", cfg.PageRange, "1-4")
	}
	if cfg.Quality != "high" {
		t.Errorf("Quality = %q, want %q", cfg.Quality, "high")
	}
	if cfg.Duplex != "long-edge" {
		t.Errorf("Duplex = %q, want %q", cfg.Duplex, "long-edge")
	}
	if cfg.ColorMode != "grayscale" {
		t.Errorf("ColorMode = %q, want %q", cfg.ColorMode, "grayscale")
	}
}

func TestApplyEnvRejectsInvalidCopies(t *testing.T) {
	t.Setenv("PRINTMARK_COPIES", "not-a-number")

	cfg := Default()
	if err := applyEnv(&cfg); err == nil {
		t.Fatal("applyEnv: expected error for invalid PRINTMARK_COPIES, got nil")
	}
}

func TestApplyEnvColorOverride(t *testing.T) {
	t.Setenv("PRINTMARK_BOLD_COLOR", "#ff0000")

	cfg := Default()
	if err := applyEnv(&cfg); err != nil {
		t.Fatalf("applyEnv: %v", err)
	}
	if cfg.BoldColor != "#ff0000" {
		t.Errorf("BoldColor = %q, want %q", cfg.BoldColor, "#ff0000")
	}
	if cfg.ItalicColor != Default().ItalicColor {
		t.Errorf("ItalicColor = %q, want untouched default %q", cfg.ItalicColor, Default().ItalicColor)
	}
}

func TestParseHexColor(t *testing.T) {
	cases := []struct {
		in      string
		want    RGB
		wantErr bool
	}{
		{"#ff0000", RGB{255, 0, 0}, false},
		{"00ff00", RGB{0, 255, 0}, false}, // leading # is optional
		{"#0000FF", RGB{0, 0, 255}, false},
		{"#fff", RGB{}, true},      // too short
		{"#gggggg", RGB{}, true},   // not hex
		{"#12345678", RGB{}, true}, // too long
	}
	for _, tc := range cases {
		got, err := ParseHexColor(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("ParseHexColor(%q): expected error, got %+v", tc.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseHexColor(%q): unexpected error: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseHexColor(%q) = %+v, want %+v", tc.in, got, tc.want)
		}
	}
}

func TestApplyEnvBoolOverride(t *testing.T) {
	t.Setenv("PRINTMARK_SHOW_LINK_URLS", "true")

	cfg := Default()
	if err := applyEnv(&cfg); err != nil {
		t.Fatalf("applyEnv: %v", err)
	}
	if !cfg.ShowLinkURLs {
		t.Error("ShowLinkURLs = false, want true")
	}
}

func TestApplyEnvRejectsInvalidBool(t *testing.T) {
	t.Setenv("PRINTMARK_SHOW_LINK_URLS", "not-a-bool")

	cfg := Default()
	if err := applyEnv(&cfg); err == nil {
		t.Fatal("applyEnv: expected error for invalid PRINTMARK_SHOW_LINK_URLS, got nil")
	}
}

func TestEnvOverridesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(`body_size = 14.0`+"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	t.Setenv("PRINTMARK_BODY_SIZE", "20.0")

	cfg := Default()
	if err := mergeFile(&cfg, path); err != nil {
		t.Fatalf("mergeFile: %v", err)
	}
	if err := applyEnv(&cfg); err != nil {
		t.Fatalf("applyEnv: %v", err)
	}

	if cfg.BodySize != 20.0 {
		t.Errorf("BodySize = %v, want env override 20.0", cfg.BodySize)
	}
}

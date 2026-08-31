package config

import (
	"fmt"
	"strconv"
	"strings"
)

// RGB is a parsed color, each channel 0-255.
type RGB struct {
	R, G, B int
}

// ParseHexColor parses a "#RRGGBB" or "RRGGBB" hex color string.
func ParseHexColor(s string) (RGB, error) {
	h := strings.TrimPrefix(s, "#")
	if len(h) != 6 {
		return RGB{}, fmt.Errorf("invalid color %q: want 6 hex digits, e.g. \"#ff0000\"", s)
	}
	v, err := strconv.ParseUint(h, 16, 32)
	if err != nil {
		return RGB{}, fmt.Errorf("invalid color %q: %w", s, err)
	}
	return RGB{
		R: int(v >> 16 & 0xff),
		G: int(v >> 8 & 0xff),
		B: int(v & 0xff),
	}, nil
}

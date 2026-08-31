// Package printer sends files to the system's default printer.
package printer

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// Options controls how Print invokes lp. Every field's zero value means
// "don't ask for anything special" — lp/CUPS and the printer's own
// driver decide, exactly like running lp by hand with no flags.
type Options struct {
	Printer   string // -d <name>; empty uses the system default printer
	Copies    int    // -n <copies>; 0 or 1 means the default single copy
	PageRange string // -P <range>, e.g. "1-4,7,9-12"; empty means all pages
	Quality   string // "draft", "normal", "high"; empty means the printer's default
	Duplex    string // "off", "long-edge", "short-edge"; empty means the printer's default
	ColorMode string // "color", "grayscale"; empty means the printer's default
}

// Print sends the file at path to a printer via lp (CUPS), using opts to
// build its arguments. CUPS ships by default on macOS and most Linux
// distributions.
func Print(path string, opts Options) error {
	args, err := opts.lpArgs()
	if err != nil {
		return err
	}
	args = append(args, path)

	cmd := exec.Command("lp", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("lp: %w: %s", err, out)
	}
	return nil
}

// lpArgs translates Options into lp command-line arguments (everything
// except the file path itself), validating the enum-like fields along
// the way so a typo surfaces as a clear printmark error instead of a
// cryptic one from lp/CUPS.
func (o Options) lpArgs() ([]string, error) {
	var args []string

	if o.Printer != "" {
		args = append(args, "-d", o.Printer)
	}
	if o.Copies > 1 {
		args = append(args, "-n", strconv.Itoa(o.Copies))
	} else if o.Copies < 0 {
		return nil, fmt.Errorf("copies must be at least 1, got %d", o.Copies)
	}
	if o.PageRange != "" {
		args = append(args, "-P", o.PageRange)
	}

	if o.Quality != "" {
		v, err := qualityValue(o.Quality)
		if err != nil {
			return nil, err
		}
		args = append(args, "-o", "print-quality="+v)
	}
	if o.Duplex != "" {
		v, err := duplexValue(o.Duplex)
		if err != nil {
			return nil, err
		}
		args = append(args, "-o", "sides="+v)
	}
	if o.ColorMode != "" {
		v, err := colorModeValue(o.ColorMode)
		if err != nil {
			return nil, err
		}
		args = append(args, "-o", "print-color-mode="+v)
	}

	return args, nil
}

// qualityValue maps a user-facing quality name to the standard IPP
// print-quality enum value.
func qualityValue(quality string) (string, error) {
	switch strings.ToLower(quality) {
	case "draft":
		return "3", nil
	case "normal":
		return "4", nil
	case "high":
		return "5", nil
	default:
		return "", fmt.Errorf("invalid quality %q: want draft, normal, or high", quality)
	}
}

// duplexValue maps a user-facing duplex name to the standard CUPS/IPP
// "sides" option value.
func duplexValue(duplex string) (string, error) {
	switch strings.ToLower(duplex) {
	case "off":
		return "one-sided", nil
	case "long-edge":
		return "two-sided-long-edge", nil
	case "short-edge":
		return "two-sided-short-edge", nil
	default:
		return "", fmt.Errorf("invalid duplex %q: want off, long-edge, or short-edge", duplex)
	}
}

// colorModeValue maps a user-facing color mode name to the standard IPP
// print-color-mode value.
func colorModeValue(mode string) (string, error) {
	switch strings.ToLower(mode) {
	case "color":
		return "color", nil
	case "grayscale":
		return "monochrome", nil
	default:
		return "", fmt.Errorf("invalid color mode %q: want color or grayscale", mode)
	}
}

// Open launches the file at path in the OS's default viewer, for
// previewing a rendered PDF instead of sending it to a printer.
//
// Uses macOS's "open" command. Linux/Windows support (xdg-open / start)
// can be added when printmark targets those platforms.
func Open(path string) error {
	cmd := exec.Command("open", path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("open: %w: %s", err, out)
	}
	return nil
}

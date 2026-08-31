package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"

	"github.com/codysprojects/printmark/internal/config"
	"github.com/codysprojects/printmark/internal/pdfrender"
	"github.com/codysprojects/printmark/internal/printer"
)

func main() {
	fv := registerFlags()
	pflag.Parse()

	if pflag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: printmark [flags] <path/to/file.md>")
		pflag.PrintDefaults()
		os.Exit(1)
	}
	inPath := pflag.Arg(0)

	source, err := os.ReadFile(inPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "printmark: reading %s: %v\n", inPath, err)
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "printmark: loading config: %v\n", err)
		os.Exit(1)
	}
	fv.applyTo(&cfg)

	outPath, err := tempPDFPath(inPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "printmark: %v\n", err)
		os.Exit(1)
	}

	if err := pdfrender.Render(source, outPath, cfg, filepath.Dir(inPath)); err != nil {
		fmt.Fprintf(os.Stderr, "printmark: rendering PDF: %v\n", err)
		os.Exit(1)
	}

	if *fv.preview {
		// Left for the OS's temp-file cleanup rather than removed
		// immediately: "open" hands off to the viewer asynchronously,
		// so deleting the file here could race its own read of it.
		if err := printer.Open(outPath); err != nil {
			fmt.Fprintf(os.Stderr, "printmark: opening preview: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "printmark: preview written to %s\n", outPath)
		return
	}

	defer os.Remove(outPath)

	popts := printer.Options{
		Printer:   cfg.Printer,
		Copies:    cfg.Copies,
		PageRange: cfg.PageRange,
		Quality:   cfg.Quality,
		Duplex:    cfg.Duplex,
		ColorMode: cfg.ColorMode,
	}
	if err := printer.Print(outPath, popts); err != nil {
		fmt.Fprintf(os.Stderr, "printmark: printing: %v\n", err)
		os.Exit(1)
	}
}

func tempPDFPath(inPath string) (string, error) {
	f, err := os.CreateTemp("", filepath.Base(inPath)+"-*.pdf")
	if err != nil {
		return "", err
	}
	name := f.Name()
	f.Close()
	return name, nil
}

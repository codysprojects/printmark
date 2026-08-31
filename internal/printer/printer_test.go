package printer

import (
	"reflect"
	"testing"
)

func TestLpArgsAllZeroValueMeansNoExtraFlags(t *testing.T) {
	args, err := Options{}.lpArgs()
	if err != nil {
		t.Fatalf("lpArgs: %v", err)
	}
	if len(args) != 0 {
		t.Errorf("lpArgs = %v, want no arguments for all-zero Options", args)
	}
}

func TestLpArgsBuildsExpectedFlags(t *testing.T) {
	opts := Options{
		Printer:   "MyPrinter",
		Copies:    3,
		PageRange: "1-4,7",
		Quality:   "high",
		Duplex:    "long-edge",
		ColorMode: "grayscale",
	}
	want := []string{
		"-d", "MyPrinter",
		"-n", "3",
		"-P", "1-4,7",
		"-o", "print-quality=5",
		"-o", "sides=two-sided-long-edge",
		"-o", "print-color-mode=monochrome",
	}

	got, err := opts.lpArgs()
	if err != nil {
		t.Fatalf("lpArgs: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("lpArgs = %v, want %v", got, want)
	}
}

func TestLpArgsSingleCopyOmitsFlag(t *testing.T) {
	args, err := Options{Copies: 1}.lpArgs()
	if err != nil {
		t.Fatalf("lpArgs: %v", err)
	}
	if len(args) != 0 {
		t.Errorf("lpArgs = %v, want no -n flag for a single copy", args)
	}
}

func TestLpArgsRejectsNegativeCopies(t *testing.T) {
	if _, err := (Options{Copies: -1}).lpArgs(); err == nil {
		t.Fatal("lpArgs: expected error for negative copies, got nil")
	}
}

func TestLpArgsRejectsInvalidQuality(t *testing.T) {
	if _, err := (Options{Quality: "ultra"}).lpArgs(); err == nil {
		t.Fatal("lpArgs: expected error for invalid quality, got nil")
	}
}

func TestLpArgsRejectsInvalidDuplex(t *testing.T) {
	if _, err := (Options{Duplex: "sideways"}).lpArgs(); err == nil {
		t.Fatal("lpArgs: expected error for invalid duplex, got nil")
	}
}

func TestLpArgsRejectsInvalidColorMode(t *testing.T) {
	if _, err := (Options{ColorMode: "sepia"}).lpArgs(); err == nil {
		t.Fatal("lpArgs: expected error for invalid color mode, got nil")
	}
}

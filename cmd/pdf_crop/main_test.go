package main

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestParseArgs_Help(t *testing.T) {
	_, err := parseArgs([]string{"--help"})
	if !errors.Is(err, errHelp) {
		t.Fatalf("expected help error, got %v", err)
	}
}

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	f()
	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stdout = old
	return buf.String()
}

func TestPrintUsage_ContainsKeyLines(t *testing.T) {
	out := captureOutput(func() { printUsage() })
	checks := []string{"pdf_crop - Crop PDF pages", "Usage:", "--threshold", "--space", "--dpi", "--help"}
	for _, c := range checks {
		if !strings.Contains(out, c) {
			t.Fatalf("usage missing %q: %s", c, out)
		}
	}
}

func TestParseArgs_MissingInput(t *testing.T) {
	_, err := parseArgs([]string{"--threshold", "0.01"})
	if err == nil {
		t.Fatalf("expected error for missing input")
	}
}

func TestParseArgs_UnknownArg(t *testing.T) {
	_, err := parseArgs([]string{"--unknown"})
	if err == nil {
		t.Fatalf("expected error for unknown arg")
	}
}

func TestParseArgs_PageNeedsSixArgs(t *testing.T) {
	_, err := parseArgs([]string{"-i", "in.pdf", "-p", "1", "10"})
	if err == nil {
		t.Fatalf("expected error for insufficient --page args")
	}
}

func TestParseArgs_ValidMultiplePages(t *testing.T) {
	args, err := parseArgs([]string{
		"-i", "in.pdf",
		"--threshold", "0.02",
		"--space", "7",
		"--dpi", "200",
		"-p", "0", "10", "10", "100", "100", "out0.pdf",
		"-p", "1", "0", "0", "0", "0", "out1.pdf",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if args.InputFile != "in.pdf" || args.Space != 7 || args.DPI != 200 || args.Threshold != 0.02 {
		t.Fatalf("parsed values unexpected: %+v", args)
	}
	if len(args.Pages) != 2 {
		t.Fatalf("expected 2 pages, got %d", len(args.Pages))
	}
}

func TestParseArgs_InvalidDPI(t *testing.T) {
	_, err := parseArgs([]string{"-i", "in.pdf", "--dpi", "abc"})
	if err == nil {
		t.Fatalf("expected error for invalid dpi")
	}
}

func TestParseArgs_InvalidSpace(t *testing.T) {
	_, err := parseArgs([]string{"-i", "in.pdf", "--space", "nope"})
	if err == nil {
		t.Fatalf("expected error for invalid space")
	}
}

func TestParseArgs_MissingValueForThreshold(t *testing.T) {
	_, err := parseArgs([]string{"-i", "in.pdf", "--threshold"})
	if err == nil {
		t.Fatalf("expected error for missing threshold value")
	}
}

func TestParseArgs_HelpShortFlag(t *testing.T) {
	_, err := parseArgs([]string{"-h"})
	if !errors.Is(err, errHelp) {
		t.Fatalf("expected help error for -h")
	}
}

func TestParseArgs_MissingValueForDPI(t *testing.T) {
	_, err := parseArgs([]string{"-i", "in.pdf", "--dpi"})
	if err == nil {
		t.Fatalf("expected error for missing dpi value")
	}
}

func TestParseArgs_MissingValueForSpace(t *testing.T) {
	_, err := parseArgs([]string{"-i", "in.pdf", "--space"})
	if err == nil {
		t.Fatalf("expected error for missing space value")
	}
}

func TestParseArgs_MissingValueForInput(t *testing.T) {
	_, err := parseArgs([]string{"-i"})
	if err == nil {
		t.Fatalf("expected error for missing input value")
	}
}

func TestParseArgs_InvalidPageNumber(t *testing.T) {
	_, err := parseArgs([]string{"-i", "in.pdf", "-p", "x", "10", "10", "100", "100", "out.pdf"})
	if err == nil {
		t.Fatalf("expected error for invalid page number")
	}
}

func TestParseArgs_InvalidLeftValue(t *testing.T) {
	_, err := parseArgs([]string{"-i", "in.pdf", "-p", "0", "x", "10", "100", "100", "out.pdf"})
	if err == nil {
		t.Fatalf("expected error for invalid left value")
	}
}

func TestParseArgs_InvalidTopValue(t *testing.T) {
	_, err := parseArgs([]string{"-i", "in.pdf", "-p", "0", "10", "x", "100", "100", "out.pdf"})
	if err == nil {
		t.Fatalf("expected error for invalid top value")
	}
}

func TestParseArgs_InvalidRightValue(t *testing.T) {
	_, err := parseArgs([]string{"-i", "in.pdf", "-p", "0", "10", "10", "x", "100", "out.pdf"})
	if err == nil {
		t.Fatalf("expected error for invalid right value")
	}
}

func TestParseArgs_InvalidBottomValue(t *testing.T) {
	_, err := parseArgs([]string{"-i", "in.pdf", "-p", "0", "10", "10", "100", "x", "out.pdf"})
	if err == nil {
		t.Fatalf("expected error for invalid bottom value")
	}
}

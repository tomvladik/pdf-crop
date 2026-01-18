package main

import (
	"errors"
	"bytes"
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

func TestParseArgs_UnknownArg(t *testing.T) {
	_, err := parseArgs([]string{"--unknown"})
	if err == nil {
		t.Fatalf("expected error for unknown arg")
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
	checks := []string{"crop_all_pdf - Crop", "Usage:", "--threshold", "--space", "--dpi", "--help"}
	for _, c := range checks {
		if !strings.Contains(out, c) {
			t.Fatalf("usage missing %q: %s", c, out)
		}
	}
}

func TestParseArgs_Valid(t *testing.T) {
	args, err := parseArgs([]string{
		"--dir", "/tmp",
		"--threshold", "0.02",
		"--space", "9",
		"--dpi", "150",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if args.Dir != "/tmp" || args.Space != 9 || args.DPI != 150 || args.Threshold != 0.02 {
		t.Fatalf("parsed values unexpected: %+v", args)
	}
}

func TestParseArgs_MissingDirValue(t *testing.T) {
	_, err := parseArgs([]string{"--dir"})
	if err == nil {
		t.Fatalf("expected error for missing dir value")
	}
}

func TestParseArgs_InvalidThreshold(t *testing.T) {
	_, err := parseArgs([]string{"--dir", "/tmp", "--threshold", "x"})
	if err == nil {
		t.Fatalf("expected error for invalid threshold")
	}
}

func TestParseArgs_InvalidDPI(t *testing.T) {
	_, err := parseArgs([]string{"--dir", "/tmp", "--dpi", "x"})
	if err == nil {
		t.Fatalf("expected error for invalid dpi")
	}
}

func TestParseArgs_InvalidSpace(t *testing.T) {
	_, err := parseArgs([]string{"--dir", "/tmp", "--space", "x"})
	if err == nil {
		t.Fatalf("expected error for invalid space")
	}
}

func TestParseArgs_MissingValueForThreshold(t *testing.T) {
	_, err := parseArgs([]string{"--dir", "/tmp", "--threshold"})
	if err == nil {
		t.Fatalf("expected error for missing threshold value")
	}
}

func TestParseArgs_MissingValueForDPI(t *testing.T) {
	_, err := parseArgs([]string{"--dir", "/tmp", "--dpi"})
	if err == nil {
		t.Fatalf("expected error for missing dpi value")
	}
}

func TestParseArgs_MissingValueForSpace(t *testing.T) {
	_, err := parseArgs([]string{"--dir", "/tmp", "--space"})
	if err == nil {
		t.Fatalf("expected error for missing space value")
	}
}

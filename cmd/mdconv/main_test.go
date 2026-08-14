package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunReadsFromStdinAndWritesToStdout(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run(nil, strings.NewReader("# Title"), &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", exitCode, stderr.String())
	}
	if stdout.String() != "<h1>Title</h1>" {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}
}

func TestRunReadsInputFile(t *testing.T) {
	inputPath := filepath.Join(t.TempDir(), "input.md")
	if err := os.WriteFile(inputPath, []byte("Paragraph"), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{inputPath}, strings.NewReader("ignored"), &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", exitCode, stderr.String())
	}
	if stdout.String() != "<p>Paragraph</p>" {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}
}

func TestRunWritesOutputFileWhenRequested(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "output.html")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"-o", outputPath}, strings.NewReader("# Title"), &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", exitCode, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "<h1>Title</h1>" {
		t.Fatalf("unexpected output file content: %q", content)
	}
}

func TestRunSupportsFlagsAfterInputFile(t *testing.T) {
	inputPath := filepath.Join(t.TempDir(), "input.md")
	outputPath := filepath.Join(t.TempDir(), "output.html")
	if err := os.WriteFile(inputPath, []byte("# Title"), 0o600); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{inputPath, "-o", outputPath, "--to", "html"}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
}

func TestRunReportsMissingInputFile(t *testing.T) {
	var stderr bytes.Buffer

	exitCode := run([]string{"missing.md"}, strings.NewReader(""), &bytes.Buffer{}, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "read missing.md") {
		t.Fatalf("unexpected stderr: %q", stderr.String())
	}
}

func TestRunRejectsUnsupportedOutputFormat(t *testing.T) {
	var stderr bytes.Buffer

	exitCode := run([]string{"--to", "text"}, strings.NewReader("content"), &bytes.Buffer{}, &stderr)

	if exitCode != 2 {
		t.Fatalf("expected exit code 2, got %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "unsupported output format") {
		t.Fatalf("unexpected stderr: %q", stderr.String())
	}
}

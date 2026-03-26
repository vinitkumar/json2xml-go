package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"-h"}, strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stderr.String(), "json2xml-go - Convert JSON to XML") {
		t.Errorf("expected usage in stderr, got %s", stderr.String())
	}
}

func TestRunVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"-v"}, strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout.String(), "json2xml-go version") {
		t.Errorf("expected version in stdout, got %s", stdout.String())
	}
}

func TestRunFromString(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"-s", `{"name":"test"}`, "-p=false", "-t=false"}, strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "<name>test</name>") {
		t.Errorf("expected XML output, got %s", stdout.String())
	}
}

func TestRunFromStdin(t *testing.T) {
	var stdout, stderr bytes.Buffer
	stdin := strings.NewReader(`{"key":"value"}`)
	code := run([]string{"-p=false", "-t=false", "-"}, stdin, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "<key>value</key>") {
		t.Errorf("expected XML output, got %s", stdout.String())
	}
}

func TestRunFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.json")
	if err := os.WriteFile(tmpFile, []byte(`{"hello":"world"}`), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := run([]string{"-p=false", "-t=false", tmpFile}, strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "<hello>world</hello>") {
		t.Errorf("expected XML output, got %s", stdout.String())
	}
}

func TestRunOutputToFile(t *testing.T) {
	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "out.xml")

	var stdout, stderr bytes.Buffer
	code := run([]string{"-s", `{"a":"b"}`, "-o", outFile, "-p=false", "-t=false"}, strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, stderr.String())
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if !strings.Contains(string(data), "<a>b</a>") {
		t.Errorf("expected XML in output file, got %s", string(data))
	}
}

func TestRunNoInput(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{}, strings.NewReader(""), &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "no input provided") {
		t.Errorf("expected error message, got %s", stderr.String())
	}
}

func TestRunInvalidJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"-s", `{broken`}, strings.NewReader(""), &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "Error reading input") {
		t.Errorf("expected error message, got %s", stderr.String())
	}
}

func TestRunXPathFormat(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"-s", `{"n":1}`, "-x", "-p=false"}, strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "xpath-functions") {
		t.Errorf("expected XPath format, got %s", stdout.String())
	}
}

func TestRunCustomWrapper(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"-s", `{"k":"v"}`, "-w", "myroot", "-p=false", "-t=false"}, strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "<myroot>") {
		t.Errorf("expected custom wrapper, got %s", stdout.String())
	}
}

func TestRunEmptyStdin(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"-"}, strings.NewReader(""), &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "empty input") {
		t.Errorf("expected empty input error, got %s", stderr.String())
	}
}

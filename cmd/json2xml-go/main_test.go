package main

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type cliState struct {
	inputURL    string
	inputString string
	outputFile  string
	wrapper     string
	root        bool
	pretty      bool
	attrType    bool
	itemWrap    bool
	xpathFormat bool
	cdata       bool
	listHeaders bool
	showVersion bool
	showHelp    bool
	stdin       *os.File
}

func saveCLIState(t *testing.T) cliState {
	t.Helper()
	state := cliState{
		inputURL:    inputURL,
		inputString: inputString,
		outputFile:  outputFile,
		wrapper:     wrapper,
		root:        root,
		pretty:      pretty,
		attrType:    attrType,
		itemWrap:    itemWrap,
		xpathFormat: xpathFormat,
		cdata:       cdata,
		listHeaders: listHeaders,
		showVersion: showVersion,
		showHelp:    showHelp,
		stdin:       os.Stdin,
	}
	t.Cleanup(func() {
		inputURL = state.inputURL
		inputString = state.inputString
		outputFile = state.outputFile
		wrapper = state.wrapper
		root = state.root
		pretty = state.pretty
		attrType = state.attrType
		itemWrap = state.itemWrap
		xpathFormat = state.xpathFormat
		cdata = state.cdata
		listHeaders = state.listHeaders
		showVersion = state.showVersion
		showHelp = state.showHelp
		os.Stdin = state.stdin
		if err := flag.CommandLine.Parse([]string{}); err != nil {
			t.Fatalf("failed to reset flags: %v", err)
		}
	})

	inputURL = ""
	inputString = ""
	outputFile = ""
	wrapper = "all"
	root = true
	pretty = true
	attrType = true
	itemWrap = true
	xpathFormat = false
	cdata = false
	listHeaders = false
	showVersion = false
	showHelp = false
	if err := flag.CommandLine.Parse([]string{}); err != nil {
		t.Fatalf("failed to clear flags: %v", err)
	}

	return state
}

func TestUsageWritesHelp(t *testing.T) {
	saveCLIState(t)

	var stderr bytes.Buffer
	usageTo(&stderr)

	output := stderr.String()
	for _, want := range []string{"json2xml-go - Convert JSON to XML", "Input Options:", "--xpath"} {
		if !strings.Contains(output, want) {
			t.Fatalf("usage output missing %q: %s", want, output)
		}
	}
}

func TestUsageWritesToStderr(t *testing.T) {
	saveCLIState(t)
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	originalStderr := os.Stderr
	os.Stderr = writer
	t.Cleanup(func() {
		os.Stderr = originalStderr
		if err := reader.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
			t.Fatalf("failed to close reader: %v", err)
		}
	})

	usage()

	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}
	contents, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read stderr: %v", err)
	}
	if !strings.Contains(string(contents), "json2xml-go - Convert JSON to XML") {
		t.Fatalf("expected usage output on stderr, got %q", contents)
	}
}

func TestRunHelp(t *testing.T) {
	saveCLIState(t)
	showHelp = true

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := run(&stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "Usage:") {
		t.Fatalf("expected help on stderr, got %q", stderr.String())
	}
}

func TestRunVersion(t *testing.T) {
	saveCLIState(t)
	showVersion = true

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := run(&stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr, got %q", stderr.String())
	}
	for _, want := range []string{"json2xml-go version", "Author:"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("version output missing %q: %s", want, stdout.String())
		}
	}
}

func TestRunConvertsStringToStdout(t *testing.T) {
	saveCLIState(t)
	inputString = `{"name":"Bike","active":true}`
	wrapper = "bike"
	pretty = false
	attrType = false

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := run(&stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", exitCode, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr, got %q", stderr.String())
	}
	output := stdout.String()
	for _, want := range []string{"<bike>", "<name>Bike</name>", "<active>true</active>"} {
		if !strings.Contains(output, want) {
			t.Fatalf("output missing %q: %s", want, output)
		}
	}
}

func TestRunReportsReadErrors(t *testing.T) {
	saveCLIState(t)
	inputString = `{not json}`

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := run(&stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "Error reading input:") {
		t.Fatalf("expected read error on stderr, got %q", stderr.String())
	}
}

func TestRunReportsConversionErrors(t *testing.T) {
	saveCLIState(t)
	inputString = `{"name":"Bike"}`
	wrapper = "bad<root"

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := run(&stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "Error converting to XML:") {
		t.Fatalf("expected conversion error on stderr, got %q", stderr.String())
	}
}

func TestRunReportsWriteErrors(t *testing.T) {
	saveCLIState(t)
	inputString = `{"name":"Bike"}`
	outputFile = t.TempDir()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := run(&stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "Error writing output:") {
		t.Fatalf("expected write error on stderr, got %q", stderr.String())
	}
}

func TestReadInputFromString(t *testing.T) {
	saveCLIState(t)
	inputString = `{"name":"Bike","active":true}`

	data, err := readInput()
	if err != nil {
		t.Fatalf("readInput returned error: %v", err)
	}

	expected := map[string]any{"name": "Bike", "active": true}
	if !reflect.DeepEqual(data, expected) {
		t.Fatalf("expected %#v, got %#v", expected, data)
	}
}

func TestReadInputFromURL(t *testing.T) {
	saveCLIState(t)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if _, err := io.WriteString(writer, `{"name":"Bike"}`); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	t.Cleanup(server.Close)
	inputURL = server.URL

	data, err := readInput()
	if err != nil {
		t.Fatalf("readInput returned error: %v", err)
	}

	expected := map[string]any{"name": "Bike"}
	if !reflect.DeepEqual(data, expected) {
		t.Fatalf("expected %#v, got %#v", expected, data)
	}
}

func TestReadInputFromFile(t *testing.T) {
	saveCLIState(t)
	inputFile := filepath.Join(t.TempDir(), "input.json")
	if err := os.WriteFile(inputFile, []byte(`{"name":"Bike"}`), 0644); err != nil {
		t.Fatalf("failed to write input file: %v", err)
	}
	if err := flag.CommandLine.Parse([]string{inputFile}); err != nil {
		t.Fatalf("failed to parse flags: %v", err)
	}

	data, err := readInput()
	if err != nil {
		t.Fatalf("readInput returned error: %v", err)
	}

	expected := map[string]any{"name": "Bike"}
	if !reflect.DeepEqual(data, expected) {
		t.Fatalf("expected %#v, got %#v", expected, data)
	}
}

func TestReadInputFromStdinArg(t *testing.T) {
	saveCLIState(t)
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdin = reader
	t.Cleanup(func() {
		if err := reader.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
			t.Fatalf("failed to close reader: %v", err)
		}
	})

	if _, err := io.WriteString(writer, `{"name":"Bike"}`); err != nil {
		t.Fatalf("failed to write stdin data: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}
	if err := flag.CommandLine.Parse([]string{"-"}); err != nil {
		t.Fatalf("failed to parse flags: %v", err)
	}

	data, err := readInput()
	if err != nil {
		t.Fatalf("readInput returned error: %v", err)
	}

	expected := map[string]any{"name": "Bike"}
	if !reflect.DeepEqual(data, expected) {
		t.Fatalf("expected %#v, got %#v", expected, data)
	}
}

func TestReadInputFromImplicitStdin(t *testing.T) {
	saveCLIState(t)
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdin = reader
	t.Cleanup(func() {
		if err := reader.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
			t.Fatalf("failed to close reader: %v", err)
		}
	})

	if _, err := io.WriteString(writer, `{"name":"Bike"}`); err != nil {
		t.Fatalf("failed to write stdin data: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	data, err := readInput()
	if err != nil {
		t.Fatalf("readInput returned error: %v", err)
	}

	expected := map[string]any{"name": "Bike"}
	if !reflect.DeepEqual(data, expected) {
		t.Fatalf("expected %#v, got %#v", expected, data)
	}
}

func TestReadInputWithoutInput(t *testing.T) {
	saveCLIState(t)
	devNull, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("failed to open null device: %v", err)
	}
	os.Stdin = devNull
	t.Cleanup(func() {
		if err := devNull.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
			t.Fatalf("failed to close null device: %v", err)
		}
	})

	_, err = readInput()
	if err == nil || !strings.Contains(err.Error(), "no input provided") {
		t.Fatalf("expected no input error, got %v", err)
	}
}

func TestReadInputReportsStdinStatError(t *testing.T) {
	saveCLIState(t)
	reader, _, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdin = reader
	if err := reader.Close(); err != nil {
		t.Fatalf("failed to close reader: %v", err)
	}

	_, err = readInput()
	if err == nil || !strings.Contains(err.Error(), "failed to inspect stdin") {
		t.Fatalf("expected stdin stat error, got %v", err)
	}
}

func TestReadFromStdinRejectsEmptyInput(t *testing.T) {
	saveCLIState(t)
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdin = reader
	t.Cleanup(func() {
		if err := reader.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
			t.Fatalf("failed to close reader: %v", err)
		}
	})
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	_, err = readFromStdin()
	if err == nil || !strings.Contains(err.Error(), "empty input") {
		t.Fatalf("expected empty input error, got %v", err)
	}
}

func TestReadFromStdinReportsReadError(t *testing.T) {
	saveCLIState(t)
	reader, _, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdin = reader
	if err := reader.Close(); err != nil {
		t.Fatalf("failed to close reader: %v", err)
	}

	_, err = readFromStdin()
	if err == nil || !strings.Contains(err.Error(), "failed to read from stdin") {
		t.Fatalf("expected stdin read error, got %v", err)
	}
}

func TestWriteOutputToFile(t *testing.T) {
	saveCLIState(t)
	outputFile = filepath.Join(t.TempDir(), "output.xml")
	if err := writeOutput("<root></root>"); err != nil {
		t.Fatalf("writeOutput returned error: %v", err)
	}

	contents, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if string(contents) != "<root></root>" {
		t.Fatalf("expected output file contents %q, got %q", "<root></root>", contents)
	}
}

func TestWriteOutputToWriter(t *testing.T) {
	saveCLIState(t)

	var stdout bytes.Buffer
	if err := writeOutputTo(&stdout, "<root></root>"); err != nil {
		t.Fatalf("writeOutputTo returned error: %v", err)
	}

	if stdout.String() != "<root></root>\n" {
		t.Fatalf("expected stdout contents %q, got %q", "<root></root>\n", stdout.String())
	}
}

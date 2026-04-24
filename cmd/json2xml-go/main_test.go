package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestReadInputFromString(t *testing.T) {
	originalInputString := inputString
	originalInputURL := inputURL
	defer func() {
		inputString = originalInputString
		inputURL = originalInputURL
	}()

	inputURL = ""
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

func TestWriteOutputToFile(t *testing.T) {
	originalOutputFile := outputFile
	defer func() { outputFile = originalOutputFile }()

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

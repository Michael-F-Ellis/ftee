package main

import (
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestExtractFileNames(t *testing.T) {
	// Good lines
	line := "// FTEE foo.bar baz.txt"
	expected := []string{"foo.bar", "baz.txt"}
	names, err := extractFileNames("FTEE", line)
	if err != nil {
		t.Errorf("Unexpected error \"%q\" parsing \"%s\"", err, line)
	}
	if !reflect.DeepEqual(names, expected) {
		t.Errorf("Expected %s got %s", expected, names)
	}
	// Output lines (no FTEE)
	line = "lorem ipsum sit amet ..."
	names, err = extractFileNames("FTEE", line)
	if err != nil {
		t.Errorf("Unexpected error \"%q\" parsing \"%s\"", err, line)
	}

	// Bad lines
	line = "//FTEE foo.bar baz.txt"
	errexp := fmt.Errorf("Delimiter FTEE must be surrounded by whitespace")
	names, err = extractFileNames("FTEE", line)
	if !reflect.DeepEqual(err, errexp) {
		t.Errorf("Expected %q got %q", errexp, err)
	}
	line = "// FTEE foo.bar FTEE baz.txt"
	errexp = fmt.Errorf("Found more than one delimiter FTEE in line.")
	names, err = extractFileNames("FTEE", line)
	if !reflect.DeepEqual(err, errexp) {
		t.Errorf("Expected %q got %q", errexp, err)
	}
	line = "// FTEE"
	errexp = fmt.Errorf("No file names found after delimiter FTEE")
	names, err = extractFileNames("FTEE", line)
	if !reflect.DeepEqual(err, errexp) {
		t.Errorf("Expected %q got %q", errexp, err)
	}

}

func TestOpenOutputFiles(t *testing.T) {
	//outputs := make(map[string]*os.File)
	names := []string{"/tmp/foo.txt", "/tmp/bar.txt"}
	err := openOutputFiles(names)
	if err != nil {
		t.Errorf("Unexpected error: %q", err)
	}
	if len(_gOutputs) != 2 {
		t.Errorf("Expected 2 opened files, got %d", len(_gOutputs))
	}
	err = openOutputFiles(names)
	if err != nil {
		t.Errorf("Unexpected error: %q", err)
	}
	if len(_gOutputs) != 2 {
		t.Errorf("Expected 2 opened files, got %d", len(_gOutputs))
	}
	closeOutputFiles()
	removeOutputFiles()
}

func BenchmarkProcessInputFile(b *testing.B) {
	// includes file I/O
	for n := 0; n < b.N; n++ {
		infd, err := os.Open("bigfile.txt")
		if err != nil {
			err = fmt.Errorf("Couldn't open input file: %q", err)
			return
		}
		processInputFile(infd, "FTEE")
		infd.Close()
		removeOutputFiles()
	}
}

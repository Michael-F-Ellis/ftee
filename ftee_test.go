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
	outputs := make(map[string]*os.File)
	names := []string{"/tmp/foo.txt", "/tmp/bar.txt"}
	updated, err := openOutputFiles(names, outputs)
	if err != nil {
		t.Errorf("Unexpected error: %q", err)
	}
	if len(updated) != 2 {
		t.Errorf("Expected 2 opened files, got %d", len(updated))
	}
	updated2, err := openOutputFiles(names, updated)
	if err != nil {
		t.Errorf("Unexpected error: %q", err)
	}
	if !reflect.DeepEqual(updated, updated2) {
		t.Errorf("Expected %v got %v", updated, updated2)
	}
	closeOutputFiles()
	removeOutputFiles()
}

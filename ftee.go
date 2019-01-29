package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var outputs = make(map[string]*os.File)

func main() {
	// Ensure we exit with an error code and log message
	// when needed after deferred cleanups have run.
	// Credit: https://tinyurl.com/ycv9zpbn
	var err error
	defer func() {
		if err != nil {
			removeOutputFiles()
			log.Fatalln(err)
		}
	}()
	defer closeOutputFiles()

	// Parse command line
	var delimiter string
	flag.StringVar(&delimiter, "d", "FTEE", "the delimiter tag")
	flag.Parse()
	infiles := flag.Args()

	var infd *os.File
	for _, infname := range infiles {
		// Open the input file
		infd, err = os.Open(infname)
		if err != nil {
			err = fmt.Errorf("Couldn't open input file: %q", err)
			return
		}
		outputs, err = processInputFile(infd, delimiter, outputs)
		if err != nil {
			err = fmt.Errorf("Error processing %s: %q", infname, err)
			return
		}
	}
}

// extractFileNames parses a line of text. If the line doesn't contain the
// delimiter, it returns an empty slice and a nil error to indicate that this
// line is to be output to whatever file targets are currently in effect.
// Otherwise it splits the line on whitespace. Each field after the delimiter
// is presumed to to be a file name and is appended to the names slice. Non-nil
// errors are returned unless the delimiter is found in exactly one field and
// there is at least on field following it.
func extractFileNames(delimiter string, line string) (names []string, err error) {
	// Short circuit if line doesn't contain delimiter
	if !strings.Contains(line, delimiter) {
		return
	}
	fields := strings.Fields(line)
	dfound := false
	for _, field := range fields {
		if !dfound {
			if field == delimiter {
				dfound = true
			}
			continue
		}
		if field == delimiter {
			err = fmt.Errorf("Found more than one delimiter %s in line.", delimiter)
			return names, err
		}
		names = append(names, field)
	}
	switch dfound {
	case false:
		err = fmt.Errorf("Delimiter %s must be surrounded by whitespace", delimiter)
	case true:
		if len(names) == 0 {
			err = fmt.Errorf("No file names found after delimiter %s", delimiter)
		}
	}
	return names, err
}

// openOutputFiles is called with results from extractFileNames. For each name
// in the list, It checks the outputs map to see if the file is already opened.
// If so, it ignores the name and moves on to the next one.  Otherwise it
// attempts to open the file for writing, truncating it if it exists. If
// successful it adds it to outputs map. On failure, it returns the error from
// os.Create immediately without attempting to open any further files from the
// names list.
func openOutputFiles(names []string, outputs map[string]*os.File) (map[string]*os.File, error) {
	var err error = nil
	for _, name := range names {
		isnew := true
		for oname, _ := range outputs {
			if oname == name {
				isnew = false
				break
			}
		}
		if isnew {
			fd, err := os.Create(name)
			if err != nil {
				return outputs, err
			}
			outputs[name] = fd
		}
	}
	return outputs, err
}

// closeOutputFiles is used as a deferred call in main to ensure that all
// output files are closed on exit.
func closeOutputFiles() {
	for _, fd := range outputs {
		fd.Close()
	}
}

// removeOutputFiles is used in main to ensure that all output files are
// removed if an error has occurred.
func removeOutputFiles() {
	for name, _ := range outputs {
		os.Remove(name)
	}
}

func processInputFile(fd *os.File, delimiter string, outputs map[string]*os.File) (map[string]*os.File, error) {
	var err error = nil
	reader := bufio.NewReader(fd)
	var targets = make([]*os.File, 0)
	var line string
	for {
		line, err = reader.ReadString('\n')
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			return outputs, err
		}
		names, err := extractFileNames(delimiter, line)
		if err != nil {
			return outputs, err
		}
		if len(names) == 0 {
			// lineout := line + "\n"
			for _, f := range targets {
				f.WriteString(line)
			}
		} else {
			outputs, err = openOutputFiles(names, outputs)
			if err != nil {
				return outputs, err
			}
			targets = make([]*os.File, 0)
			for _, name := range names {
				targets = append(targets, outputs[name])
			}
		}
	}
	return outputs, err
}

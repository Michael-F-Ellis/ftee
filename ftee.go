// Copyright 2019 Ellis & Grant, Inc. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

/*
ftee is a many-to-many file splitter. Command line usage is

	ftee [-h] [-d delimiter] infile1 [infile2 ... ]

The default delimiter is "FTEE".
*/
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

const copyright = `
Copyright 2019 Ellis & Grant, Inc. All rights reserved.  Use of the source
code is governed by an MIT-style license that can be found in the LICENSE
file.`
const description = `
  ftee reads all lines in each input file sequentially. When it sees a line
  ending with "delimiter outfile1 [outfile2 ...]", it opens the outfiles and
  writes all following lines to each output file until another delimiter line
  is encountered.
  
  EXAMPLE 
    Consider a file containing:

	  This is ignored
	  FTEE /tmp/out1
	  This goes into out1 only.
	  FTEE /tmp/out2
	  This goes into out2 only.
	  FTEE /tmp/out1 /tmp/out3
	  This goes into out1 and out3.
  
    Processing with ftee will produce 3 output files with the following content:
  
    /tmp/out1:
  	  This goes into out1 only.
  	  This goes into out1 and out3.
  
    /tmp/out2:
  	  This goes into out2 only.
  
    /tmp/out3:
  	  This goes into out1 and out3.
  
  
  ERRORS
	ftee deletes all output files and exits with an error message and a
	non-zero status code if any error occurs.

	ftee ignores the line content before the delimiter. The following are
	all correct:
    
	  FTEE somefile
	  // FTEE somefile
		# FTEE somefile
	  What do you get when cross a gopher and an elephant? An FTEE rodent
    
	The delimiter must be preceded and followed by whitespace. The following
	will cause an error:
    
	  //FTEE somefile
	  FTEEsomefile
    
	ftee expects whitespace separated valid filepaths after the delimiter to
	the end of the line. The following will cause an error:
    
	  /* FTEE somefile otherfile */

	ftee has not been tested on Windows. Problems with backslashed filepaths
	are likely.`

// Global map of output filenames and file objects.
var _gOutputs = make(map[string]*os.File)

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

	// Close any opened output files on exit.
	defer closeOutputFiles()

	// Parse command line
	flag.Usage = usage
	var delimiter string
	flag.StringVar(&delimiter, "d", "FTEE", "the delimiter tag")
	flag.Parse()
	infiles := flag.Args()

	// Process all the input files
	var infd *os.File
	for _, infname := range infiles {
		infd, err = os.Open(infname)
		if err != nil {
			err = fmt.Errorf("Couldn't open input file: %q", err)
			return
		}
		err = processInputFile(infd, delimiter)
		// Note: processInputFile handles closing the file.
		if err != nil {
			err = fmt.Errorf("Error processing %s: %q", infname, err)
			return
		}
	}
}

// usage extends the flag package's default help message.
func usage() {
	fmt.Println(copyright)
	fmt.Printf("Usage: ftee [OPTIONS] filepath [filepath ...]\n  -h    print this help message.\n")
	flag.PrintDefaults()
	fmt.Println(description)

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
func openOutputFiles(names []string) error {
	var err error = nil
	for _, name := range names {
		isnew := true
		for oname, _ := range _gOutputs {
			if oname == name {
				isnew = false
				break
			}
		}
		if isnew {
			fd, err := os.Create(name)
			if err != nil {
				return err
			}
			_gOutputs[name] = fd
		}
	}
	return err
}

// closeOutputFiles is used as a deferred call in main to ensure that all
// output files are closed on exit.
func closeOutputFiles() {
	for _, fd := range _gOutputs {
		fd.Close()
	}
}

// removeOutputFiles is used in main to ensure that all output files are
// removed if an error has occurred.
func removeOutputFiles() {
	for name, _ := range _gOutputs {
		os.Remove(name)
	}
}

// processInputFile reads every line from fd and scans to see
// if it contains delimiter. If not, the line is output to
// all currently active target files. If the line contains delimiter
// it parses the remainder of the line as a list of space-delimited
// filenames for output. These are passed to openOutputFiles() to be
// opened if they haven't been opened already. If openOutputFiles()
// succeeds, the files are set as the current output targets for
// following lines until the next delimiter line is encountered.
//
// If parsing fails, the error from extractFileNames() is returned.
// Similarly, processing ends if openOutputFiles() fails.
// Processing ends normally when all lines in the file have been
// read and processed.
func processInputFile(fd *os.File, delimiter string) error {
	defer fd.Close()
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
			return err
		}
		names, err := extractFileNames(delimiter, line)
		if err != nil {
			return err
		}
		if len(names) == 0 {
			// lineout := line + "\n"
			for _, f := range targets {
				f.WriteString(line)
			}
		} else {
			err = openOutputFiles(names)
			if err != nil {
				return err
			}
			targets = make([]*os.File, 0)
			for _, name := range names {
				targets = append(targets, _gOutputs[name])
			}
		}
	}
	return err
}

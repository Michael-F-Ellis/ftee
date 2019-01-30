# ftee
A many-to-many text file splitter written in Go

## Why
Maintaining client/server applications commonly involves editing two or more files when ever an interface is added or changed. With `ftee`, you can choose to keep the pieces together in the same file and have the makefile split the pieces apart as needed for building and testing the separate components. 

There are undoubtedly other uses but `ftee` exists because `split` and `csplit` aren't well-suited for the above.

## Installing
    go get github.com/Michael-F-Ellis/ftee
    cd go/src/github.com/Michael-F-Ellis/ftee
    go build -o /somewhere/in/your/path/ftee

## Usage
    Usage: ftee [OPTIONS] filepath [filepath ...]
      -h  print this help message.
      -d string
          the delimiter tag (default "FTEE")

    ftee reads all lines in each input file sequentially. When it sees a line
    ending with "delimiter outfile1 [outfile2 ...]", it opens the outfiles and
    writes all following lines to each output file until another delimiter line
    is encountered.

  ## Example
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


  ## Errors
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
	are likely.
	
  ## Performance
  On my 2012 Mac Mini, `ftee` processes about 65,000 lines per second including I/O to disk. See the benchmark in ftee_test.go and the comments in bigfile.txt for details.

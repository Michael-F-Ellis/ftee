# ftee
A many-to-many text file splitter written in Go

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

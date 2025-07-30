package main

import (
	"bufio"
	"io"
	"log"
	"os"
)

// ctrlDReader wraps a reader and detects Ctrl+D (ASCII 4) as EOT/EOF
type ctrlDReader struct {
	r io.Reader
}

func (c *ctrlDReader) Read(p []byte) (n int, err error) {
	n, err = c.r.Read(p)

	// Check for Ctrl+D (ASCII 4)
	for i := 0; i < n; i++ {
		if p[i] == 4 {
			// If we found Ctrl+D, return data up to that point and signal EOF
			return i, io.EOF
		}
	}

	return n, err
}

func main() {
	stdout := os.Stdout

	if len(os.Args) == 1 {
		// Create a reader that detects Ctrl+D as EOF
		stdin := &ctrlDReader{r: os.Stdin}
		reader := bufio.NewReader(stdin)

		// Copy stdin contents to stdout
		_, err := io.Copy(stdout, reader)
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
	} else {
		for _, fname := range os.Args[1:] {
			// Open file in binary mode
			fh, err := os.Open(fname)
			if err != nil {
				log.Fatal(err)
			}

			// Use buffered I/O for better performance
			reader := bufio.NewReader(fh)

			// Copy file contents preserving all bytes
			_, err = io.Copy(stdout, reader)
			if err != nil && err != io.EOF {
				log.Fatal(err)
			}

			fh.Close()
		}
	}
}

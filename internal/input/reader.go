package input

import (
	"bufio"
	"io"
	"os"
)

// Reader is an interface for reading user input
type Reader interface {
	ReadString(delim byte) (string, error)
}

// StdinReader wraps bufio.Reader for os.Stdin
type StdinReader struct {
	reader *bufio.Reader
}

// NewStdinReader creates a new StdinReader
func NewStdinReader() *StdinReader {
	return &StdinReader{
		reader: bufio.NewReader(os.Stdin),
	}
}

// ReadString reads until delimiter
func (r *StdinReader) ReadString(delim byte) (string, error) {
	return r.reader.ReadString(delim)
}

// StringReader is a simple reader for testing.
// Each input string should already include the delimiter that will be used
// in ReadString calls (e.g., "yes\n" for newline delimiter).
type StringReader struct {
	inputs []string
	index  int
}

// NewStringReader creates a reader from strings.
// Each input string should include the expected delimiter.
func NewStringReader(inputs ...string) *StringReader {
	return &StringReader{inputs: inputs}
}

// ReadString returns the next pre-configured string.
// Returns io.EOF when all inputs have been consumed.
// Note: The delim parameter is ignored; inputs should already include delimiters.
func (r *StringReader) ReadString(delim byte) (string, error) {
	if r.index >= len(r.inputs) {
		return "", io.EOF
	}
	result := r.inputs[r.index]
	r.index++
	return result, nil
}

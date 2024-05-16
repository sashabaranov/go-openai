package openai

import (
	"bufio"
	"io"
	"strings"
	"testing"
)

// NewEOLSplitterFunc returns a bufio.SplitFunc tied to a new EOLSplitter instance
func NewEOLSplitterFunc() bufio.SplitFunc {
	splitter := NewEOLSplitter()
	return splitter.Split
}

// EOLSplitter is the custom split function to handle CR LF, CR, and LF as end-of-line.
type EOLSplitter struct {
	prevCR bool
}

// NewEOLSplitter creates a new EOLSplitter instance.
func NewEOLSplitter() *EOLSplitter {
	return &EOLSplitter{prevCR: false}
}

// Split function to handle CR LF, CR, and LF as end-of-line.
func (s *EOLSplitter) Split(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Check if the previous data ended with a CR
	if s.prevCR {
		s.prevCR = false
		if len(data) > 0 && data[0] == '\n' {
			return 1, nil, nil // Skip the LF following the previous CR
		}
	}

	// Search for the first occurrence of CR LF, CR, or LF
	for i := 0; i < len(data); i++ {
		if data[i] == '\r' {
			if i+1 < len(data) && data[i+1] == '\n' {
				// Found CR LF
				return i + 2, data[:i], nil
			}
			// Found CR
			if !atEOF && i == len(data)-1 {
				// If CR is the last byte, and not EOF, then need to check if
				// the next byte is LF.
				//
				// save the state and request more data
				s.prevCR = true
				return 0, nil, nil
			}
			return i + 1, data[:i], nil
		}
		if data[i] == '\n' {
			// Found LF
			return i + 1, data[:i], nil
		}
	}

	// If at EOF, we have a final, non-terminated line. Return it.
	if atEOF && len(data) > 0 {
		return len(data), data, nil
	}

	// Request more data.
	return 0, nil, nil
}

// CustomReader simulates a reader that splits the input across multiple reads.
type CustomReader struct {
	chunks []string
	index  int
}

func NewChunksReader(chunks []string) *CustomReader {
	return &CustomReader{
		chunks: chunks,
	}
}

func (r *CustomReader) Read(p []byte) (n int, err error) {
	if r.index >= len(r.chunks) {
		return 0, io.EOF
	}
	n = copy(p, r.chunks[r.index])
	r.index++
	return n, nil
}

// TestEolSplitter tests the custom EOL splitter function.
func TestEolSplitter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"CRLF", "Line1\r\nLine2\r\nLine3\r\n", []string{"Line1", "Line2", "Line3"}},
		{"CR", "Line1\rLine2\rLine3\r", []string{"Line1", "Line2", "Line3"}},
		{"LF", "Line1\nLine2\nLine3\n", []string{"Line1", "Line2", "Line3"}},
		{"Mixed", "Line1\r\nLine2\rLine3\nLine4\r\nLine5", []string{"Line1", "Line2", "Line3", "Line4", "Line5"}},
		{"SingleLineNoEOL", "Line1", []string{"Line1"}},
		{"SingleLineLF", "Line1\n", []string{"Line1"}},
		{"SingleLineCR", "Line1\r", []string{"Line1"}},
		{"SingleLineCRLF", "Line1\r\n", []string{"Line1"}},
		{"DoubleNewLines", "Line1\n\nLine2", []string{"Line1", "", "Line2"}},
		{"lflf", "\n\n", []string{"", ""}},
		{"crlfcrlf", "\r\n\r\n", []string{"", ""}},
		{"crcr", "\r\r", []string{"", ""}},
		{"mixed eol: crlf cr lf", "A\r\nB\rC\nD", []string{"A", "B", "C", "D"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reader := strings.NewReader(test.input)
			scanner := bufio.NewScanner(reader)
			scanner.Split(NewEOLSplitterFunc())

			var lines []string
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}

			if err := scanner.Err(); err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if len(lines) != len(test.expected) {
				t.Errorf("Expected %d lines, got %d", len(test.expected), len(lines))
				t.Errorf("Expected: %v, got: %v", test.expected, lines)
			}

			for i := range lines {
				if lines[i] != test.expected[i] {
					t.Errorf("Expected line %d to be %q, got %q", i, test.expected[i], lines[i])
				}
			}
		})
	}
}

// TestEolSplitterBoundaryCondition tests the boundary condition where CR LF is split across two slices.
func TestEolSplitterBoundaryCondition(t *testing.T) {
	// Additional cases
	cases := []struct {
		input    []string
		expected []string
	}{
		{[]string{"Line1\r", "\nLine2"}, []string{"Line1", "Line2"}},
		{[]string{"Line1\r", "\nLine2\r"}, []string{"Line1", "Line2"}},
		{[]string{"Line1\r", "\nLine2\r\n"}, []string{"Line1", "Line2"}},
		{[]string{"Line1\r", "\nLine2\r", "Line3"}, []string{"Line1", "Line2", "Line3"}},
		{[]string{"Line1\r", "\nLine2\r", "\nLine3\r\n"}, []string{"Line1", "Line2", "Line3"}},
	}
	for _, c := range cases {
		// Custom reader to simulate the boundary condition
		reader := NewChunksReader(c.input)
		scanner := bufio.NewScanner(reader)
		scanner.Split(NewEOLSplitterFunc())

		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if len(lines) != len(c.expected) {
			t.Errorf("Expected %d lines, got %d", len(c.expected), len(lines))
			continue
		}

		for i := range lines {
			if lines[i] != c.expected[i] {
				t.Errorf("Expected line %d to be %q, got %q", i, c.expected[i], lines[i])
			}
		}
	}
}

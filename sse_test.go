package openai_test

import (
	"bufio"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/sashabaranov/go-openai"
)

// ChunksReader simulates a reader that splits the input across multiple reads.
type ChunksReader struct {
	chunks []string
	index  int
}

func NewChunksReader(chunks []string) *ChunksReader {
	return &ChunksReader{
		chunks: chunks,
	}
}

func (r *ChunksReader) Read(p []byte) (n int, err error) {
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
			scanner.Split(openai.NewEOLSplitterFunc())

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
		scanner.Split(openai.NewEOLSplitterFunc())

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

func TestSSEScanner(t *testing.T) {
	tests := []struct {
		raw  string
		want []openai.ServerSentEvent
	}{
		{
			raw: `data: hello world`,
			want: []openai.ServerSentEvent{
				{
					Data: "hello world",
				},
			},
		},
		{
			raw: `event: hello
data: hello world`,
			want: []openai.ServerSentEvent{
				{
					Event: "hello",
					Data:  "hello world",
				},
			},
		},
		{
			raw: `event: hello-json
data: {
data: "msg": "hello world",
data: "id": 12345
data: }`,
			want: []openai.ServerSentEvent{
				{
					Event: "hello-json",
					Data:  "{\n\"msg\": \"hello world\",\n\"id\": 12345\n}",
				},
			},
		},
		{
			raw: `data: hello world

data: hello again`,
			want: []openai.ServerSentEvent{
				{
					Data: "hello world",
				},
				{
					Data: "hello again",
				},
			},
		},
		{
			raw: `retry: 10000
			data: hello world`,
			want: []openai.ServerSentEvent{
				{
					Retry: 10000,
					Data:  "hello world",
				},
			},
		},
		{
			raw: `retry: 10000

retry: 20000`,
			want: []openai.ServerSentEvent{
				{
					Retry: 10000,
				},
				{
					Retry: 20000,
				},
			},
		},
		{
			raw: `: comment 1
: comment 2
id: message-id
retry: 20000
event: hello-event
data: hello`,
			want: []openai.ServerSentEvent{
				{
					ID:    "message-id",
					Retry: 20000,
					Event: "hello-event",
					Data:  "hello",
				},
			},
		},
		{
			raw: `: comment 1
id: message 1
data: hello 1
retry: 10000
event: hello-event 1

: comment 2
data: hello 2
id: message 2
retry: 20000
event: hello-event 2
`,
			want: []openai.ServerSentEvent{
				{
					ID:    "message 1",
					Retry: 10000,
					Event: "hello-event 1",
					Data:  "hello 1",
				},
				{
					ID:    "message 2",
					Retry: 20000,
					Event: "hello-event 2",
					Data:  "hello 2",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			rawWithCRLF := strings.ReplaceAll(tt.raw, "\n", "\r\n")
			runSSEScanTest(t, rawWithCRLF, tt.want)

			// Test with "\r" EOL
			rawWithCR := strings.ReplaceAll(tt.raw, "\n", "\r")
			runSSEScanTest(t, rawWithCR, tt.want)

			// Test with "\n" EOL (original)
			runSSEScanTest(t, tt.raw, tt.want)
		})
	}
}

func runSSEScanTest(t *testing.T, raw string, want []openai.ServerSentEvent) {
	sseScanner := openai.NewSSEScanner(strings.NewReader(raw), false)

	var got []openai.ServerSentEvent
	for sseScanner.Next() {
		got = append(got, sseScanner.Scan())
	}

	if err := sseScanner.Err(); err != nil {
		t.Errorf("SSEScanner error: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("SSEScanner() = %v, want %v", got, want)
	}
}

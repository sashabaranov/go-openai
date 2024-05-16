package openai

import (
	"bufio"
	"io"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

type ServerSentEvent struct {
	ID      string // ID of the event
	Data    string // Data of the event
	Event   string // Type of the event
	Retry   int    // Retry time in milliseconds
	Comment string // Comment
}

type SSEScanner struct {
	scanner     *bufio.Scanner
	event       *ServerSentEvent
	err         error
	readComment bool
}

func NewSSEScanner(r io.Reader, readComment bool) *SSEScanner {
	scanner := bufio.NewScanner(r)

	// N.B. The bufio.ScanLines handles `\r?\n``, but not `\r` itself as EOL, as
	// the SSE spec requires
	//
	// See: https://html.spec.whatwg.org/multipage/server-sent-events.html#parsing-an-event-stream
	//
	// scanner.Split(bufio.ScanLines)
	scanner.Split(NewEOLSplitterFunc())

	return &SSEScanner{
		scanner:     scanner,
		readComment: readComment,
	}
}

func (s *SSEScanner) Next() bool {
	s.event = nil

	var event ServerSentEvent
	var dataLines []string

	var seenNonEmptyLine bool

	for s.scanner.Scan() {
		line := strings.TrimSpace(s.scanner.Text())

		if line == "" {
			if seenNonEmptyLine {
				break
			}

			continue
		}

		seenNonEmptyLine = true

		if strings.HasPrefix(line, "id: ") {
			event.ID = strings.TrimPrefix(line, "id: ")
		} else if strings.HasPrefix(line, "data: ") {
			dataLines = append(dataLines, strings.TrimPrefix(line, "data: "))
		} else if strings.HasPrefix(line, "event: ") {
			event.Event = strings.TrimPrefix(line, "event: ")
		} else if strings.HasPrefix(line, "retry: ") {
			retry, err := strconv.Atoi(strings.TrimPrefix(line, "retry: "))
			if err == nil {
				event.Retry = retry
			}

			// ignore invalid retry values
		} else if strings.HasPrefix(line, ":") {
			if s.readComment {
				event.Comment = strings.TrimPrefix(line, ":")
			}

			// ignore comment line
		}

		// ignore unknown lines
	}

	s.err = s.scanner.Err()

	if !seenNonEmptyLine {
		return false
	}

	event.Data = strings.Join(dataLines, "\n")
	s.event = &event

	return true
}

func (s *SSEScanner) Scan() *ServerSentEvent {
	return s.event
}

func (s *SSEScanner) Err() error {
	return s.err
}

func TestSSEScanner(t *testing.T) {
	tests := []struct {
		raw  string
		want []ServerSentEvent
	}{
		{
			raw: `data: hello world`,
			want: []ServerSentEvent{
				{
					Data: "hello world",
				},
			},
		},
		{
			raw: `event: hello
data: hello world`,
			want: []ServerSentEvent{
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
			want: []ServerSentEvent{
				{
					Event: "hello-json",
					Data:  "{\n\"msg\": \"hello world\",\n\"id\": 12345\n}",
				},
			},
		},
		{
			raw: `data: hello world

data: hello again`,
			want: []ServerSentEvent{
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
			want: []ServerSentEvent{
				{
					Retry: 10000,
					Data:  "hello world",
				},
			},
		},
		{
			raw: `retry: 10000

retry: 20000`,
			want: []ServerSentEvent{
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
			want: []ServerSentEvent{
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
			want: []ServerSentEvent{
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

func runSSEScanTest(t *testing.T, raw string, want []ServerSentEvent) {
	sseScanner := NewSSEScanner(strings.NewReader(raw), false)

	var got []ServerSentEvent
	for sseScanner.Next() {
		got = append(got, *sseScanner.Scan())
	}

	if err := sseScanner.Err(); err != nil {
		t.Errorf("SSEScanner error: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("SSEScanner() = %v, want %v", got, want)
	}
}

package smeego

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
)

const (
	id_prefix    = "id"
	data_prefix  = "data"
	event_prefix = "event"
	retry_prefix = "retry"

	ping_event  = "ping"
	ready_event = "ready"
)

type sseEvent struct {
	ID    string
	Event string
	Retry string
	Data  []byte
}

type sseClient struct {
	debugWriter io.Writer
	url         string
}

func (s *sseClient) subscribeToChannel() (<-chan sseEvent, error) {
	client := http.DefaultClient

	req, err := http.NewRequest("GET", s.url, nil)
	if err != nil {
		printError(s.debugWriter, "failed to create request", err)
		return nil, ErrFailedToSubscribe
	}
	req.Header.Set("Accept", "text/event-stream")

	resp, err := client.Do(req)
	if err != nil {
		printError(s.debugWriter, "failed to subscribe", err)
		return nil, ErrFailedToSubscribe
	}

	if resp.StatusCode != http.StatusOK {
		printError(s.debugWriter, fmt.Sprintf("status code was %d", resp.StatusCode), nil)
		return nil, ErrFailedToSubscribe
	}

	if resp.Header.Get("Content-Type") != "text/event-stream" {
		printError(s.debugWriter, "content type was not text/event-stream", nil)
		return nil, ErrFailedToSubscribe
	}

	events := make(chan sseEvent)

	go s.sseNotify(resp, events)

	return events, nil
}

func (s *sseClient) sseNotify(resp *http.Response, events chan<- sseEvent) {
	ev := sseEvent{}
	var buf bytes.Buffer
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Bytes()
		printDebug(s.debugWriter, fmt.Sprintf("received line: %s", line))
		switch {
		case bytes.HasPrefix(line, []byte(id_prefix)):
			ev.ID = treatField(id_prefix, line)
		case bytes.HasPrefix(line, []byte(data_prefix)):
			buf.Write(line[len(data_prefix)+2:])
			buf.WriteByte('\n')
		case bytes.HasPrefix(line, []byte(event_prefix)):
			ev.Event = treatField(event_prefix, line)
		case bytes.HasPrefix(line, []byte(retry_prefix)):
			ev.Retry = treatField(retry_prefix, line)
		case len(line) == 0:
			ev.Data = buf.Bytes()
			events <- ev
			buf.Reset()
			ev = sseEvent{}
		default:
			printError(s.debugWriter, fmt.Sprintf("invalid event prefix: %s", line), nil)
		}

	}

	if err := scanner.Err(); err != nil {
		printError(s.debugWriter, "scanner error", err)
		close(events)
	}
}

func treatField(fieldName string, line []byte) string {
	return string(line[len(fieldName)+2:])
}

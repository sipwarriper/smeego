package smeego

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

type SmeeChannel struct {
	channelAddress string
	printer        io.Writer

	stopper chan struct{}
}

type SmeeEvent struct {
	Body      []byte
	Headers   map[string]string
	Query     []byte
	Timestamp int64
}

func NewSmeeChannel(printer io.Writer, smeeURL string) (*SmeeChannel, error) {
	httpClient := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	response, err := httpClient.Head(smeeURL + "/new")
	if err != nil {
		printError(printer, "error in head request", err)
		return nil, ErrFailedToCreateChannel
	}

	addr := response.Header.Get("Location")
	if addr == "" {
		return nil, ErrFailedToCreateChannel
	}
	return &SmeeChannel{
		channelAddress: addr,
		printer:        printer,
	}, nil
}

func (s *SmeeChannel) SubscribeToChannel() (<-chan SmeeEvent, error) {
	sseClient := sseClient{
		debugWriter: s.printer,
		url:         s.channelAddress,
	}

	events, err := sseClient.subscribeToChannel()

	if err != nil {
		return nil, err
	}

	s.stopper = make(chan struct{})
	returnChannel := make(chan SmeeEvent)

	go func() {
		for {
			select {
			case event := <-events:
				if event.Event == ping_event || event.Event == ready_event {
					continue
				}
				returnChannel <- buildSmeeEvent(event)
			case <-s.stopper:
				return
			}
		}
	}()

	return returnChannel, nil
}

func (s *SmeeChannel) Close() {
	if s.stopper != nil {
		close(s.stopper)
	}
}

func buildSmeeEvent(event sseEvent) SmeeEvent {
	middle := map[string][]byte{}
	json.Unmarshal(event.Data, &middle)

	smeeEvent := SmeeEvent{}
	for k, v := range middle {
		switch k {
		case "body":
			smeeEvent.Body = v
		case "query":
			smeeEvent.Query = v
		case "timestamp":
			val, _ := strconv.Atoi(string(v))
			smeeEvent.Timestamp = int64(val)
		default:
			smeeEvent.Headers[k] = string(v)
		}
	}
	return smeeEvent
}

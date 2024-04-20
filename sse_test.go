package smeego

import (
	"net/http"
	"strings"
	"testing"

	"github.com/sipwarriper/smeego/printers"

	"github.com/stretchr/testify/assert"
)

func TestSubscribeToChannel(t *testing.T) {
	// channel, err := NewSmeeChannel("http://localhost:3000")
	// assert.Nil(t, err)
	// assert.NotNil(t, channel)

	sseClient := sseClient{
		debugWriter: printers.SmeegoTerminalPrinter{},
		// url:         channel.channelAddress,
		url: "http://localhost:3000/azsPSntKbFg5hIeC",
	}

	events, err := sseClient.subscribeToChannel()

	assert.Nil(t, err)
	assert.NotNil(t, events)

	data := `{"event": "test"}`

	http.Post(sseClient.url, "application/json", strings.NewReader(data))

	for {
		select {
		case event := <-events:
			switch event.Event {
			case ping_event, ready_event:
				continue
			default:
				assert.Equal(t, "test", event.Event) //todo fix this

			}
		}
	}

}

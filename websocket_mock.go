package enigma

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

type (
	// MockSocket provides a dummy implementation of the Socket interface.
	MockSocket struct {
		expectedRequests chan *mocksocketRequest
		receivedMessages chan json.RawMessage
		closed           chan struct{}
	}

	mocksocketRequest struct {
		sentMessage json.RawMessage
		responses   []json.RawMessage
	}
)

func asCanonicalString(message json.RawMessage) string {
	buffer := new(bytes.Buffer)
	if err := json.Compact(buffer, message); err != nil {
		fmt.Println(err)
	}
	return buffer.String()
}

// ExpectCall sets a response message given a request message.
func (t *MockSocket) ExpectCall(request string, response string) {
	t.expectedRequests <- &mocksocketRequest{sentMessage: json.RawMessage(request), responses: []json.RawMessage{json.RawMessage(response)}}
}

// AddReceivedMessage adds a message to the received message queue immediately
func (t *MockSocket) AddReceivedMessage(response string) {
	t.receivedMessages <- json.RawMessage(response)
}

// WriteMessage implements the Socket interface
func (t *MockSocket) WriteMessage(messageType int, message []byte) error {
	select {
	case expectedMessage := <-t.expectedRequests:
		if asCanonicalString(expectedMessage.sentMessage) == asCanonicalString(message) {
			// Transfer the response into the received messages channel
			for _, response := range expectedMessage.responses {
				t.receivedMessages <- response
			}
		} else {
			fmt.Println("Unexpected response", asCanonicalString(message))
			fmt.Println("Expected ", string(expectedMessage.sentMessage))
		}
	default:
		fmt.Println("No more responses registered, expecting", string(message))
	}

	return nil
}

// ReadMessage implements the Socket interface
func (t *MockSocket) ReadMessage() (int, []byte, error) {
	message, isOpen := <-t.receivedMessages
	if !isOpen {
		return 0, nil, errors.New("socket closed by test case")
	}
	return 1, message, nil
}

// Close implements the Socket interface
func (t *MockSocket) Close() error {
	select {
	case <-t.closed:
		// Do nothing
	default:
		close(t.closed)
		close(t.expectedRequests)
		close(t.receivedMessages)
	}
	return nil
}

// NewMockSocket creates a new MockSocket instance
func NewMockSocket(fileName string) (*MockSocket, error) {
	socket := &MockSocket{receivedMessages: make(chan json.RawMessage, 100), expectedRequests: make(chan *mocksocketRequest, 10000), closed: make(chan struct{})}
	if fileName != "" {
		var lastRequest *mocksocketRequest
		messages := readTrafficLog(fileName)
		for _, m := range messages {
			if m.Sent != nil {
				lastRequest = &mocksocketRequest{sentMessage: m.Sent}
				socket.expectedRequests <- lastRequest
			} else if m.Received != nil {
				if lastRequest == nil {
					// When no request have been sent then push the received messages straight to the receivedMessage channel
					socket.receivedMessages <- m.Received
				} else {
					// When a request has been sent then add the responses to that request
					lastRequest.responses = append(lastRequest.responses, m.Received)
				}

			}
		}
	}
	return socket, nil
}

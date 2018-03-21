package enigma

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSessionEvents(t *testing.T) {

	s := newSessionEvents()
	messageChannel1 := s.SessionMessageChannel("OnConnected")
	s.emitSessionMessage("OnConnected", json.RawMessage(`{"data": "data"}`))
	messageChannel2 := s.SessionMessageChannel("OnConnected")

	data1 := <-messageChannel1
	data2 := <-messageChannel2
	assert.Equal(t, `{"data": "data"}`, string(data1.Content))
	assert.Equal(t, `{"data": "data"}`, string(data2.Content))
}

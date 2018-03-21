package enigma

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRemoteObject(t *testing.T) {

	// Create emitter, register listener
	object := newRemoteObject(nil, nil)
	channel := object.ChangedChannel()

	// Send an event
	object.signalChanged()

	// Receive the change
	<-channel

	// Unregister the listener
	object.RemoveChangeChannel(channel)
	assert.Empty(t, object.changedChannels)

	// Check that the changed channel is closed
	_, isOpen := <-channel
	assert.False(t, isOpen)

	// Check that we are still not closed
	select {
	case <-object.Closed():
		assert.Fail(t, "Should lock!")
	default:
	}

	// Close the object
	object.signalClosed()

	// Make sure we're closed now
	_, isOpen = <-object.Closed()
	assert.False(t, isOpen)

}

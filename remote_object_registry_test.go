package enigma

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRpcObjectRegistry(t *testing.T) {

	// Setup call registry"
	ror := newRemoteObjectRegistry()

	// Register object
	object := &RemoteObject{ObjectInterface: &ObjectInterface{Handle: 25}}
	ror.registerRemoteObject(object)

	// get object
	retrievedObject := ror.getRemoteObject(25)
	assert.Equal(t, retrievedObject, object)

	// UnRegister object
	unregisteredObject := ror.unregisterRemoteObject(25)
	assert.Equal(t, unregisteredObject, object)

	// UnRegistered object should be gone
	missingObject := ror.unregisterRemoteObject(25)
	assert.Nil(t, missingObject)

}

func TestRemoteObjectRegistryHandleUpdatesClosed(t *testing.T) {
	// Setup call registry"
	ror := newRemoteObjectRegistry()

	// Register object
	object := newRemoteObject(nil, &ObjectInterface{Handle: 25})
	ror.registerRemoteObject(object)

	var objectClosed bool

	// Check that the object is NOT closed
	select {
	case <-object.Closed():
		objectClosed = true
	default:
		objectClosed = false
	}
	assert.False(t, objectClosed, "The object should not be closed before the close update")

	// Send change event to handle 25
	ror.handleUpdates([]int{}, []int{25})

	// Check that the object IS closed
	select {
	case <-object.Closed():
		objectClosed = true
	default:
		objectClosed = false
	}
	assert.True(t, objectClosed, "The object should now be closed")

}

func TestRemoteObjectRegistryHandleUpdatesChanged(t *testing.T) {
	// Setup call registry"
	ror := newRemoteObjectRegistry()

	// Register object
	object := newRemoteObject(nil, &ObjectInterface{Handle: 25})
	ror.registerRemoteObject(object)

	var objectChanged bool

	changeChannel := object.ChangedChannel()
	// Check that the object is NOT closed
	select {
	case <-changeChannel:
		objectChanged = true
	default:
		objectChanged = false
	}
	assert.False(t, objectChanged, "The object should not be change before the change update")

	// Send change event to handle 25
	ror.handleUpdates([]int{}, []int{25})

	// Check that the object IS changed
	select {
	case <-changeChannel:
		objectChanged = true
	default:
		objectChanged = false
	}
	assert.True(t, objectChanged, "The object should now be changed")

	object.RemoveChangeChannel(changeChannel)
}

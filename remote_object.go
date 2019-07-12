package enigma

import (
	"context"
	"encoding/json"
	"sync"
)

type (
	// RemoteObject represents an object inside Qlik Associative Engine.
	RemoteObject struct {
		*ObjectInterface
		*session
		mutex           sync.Mutex
		changedChannels map[chan struct{}]bool
		closedCh        chan struct{}
	}
)

// ChangedChannel returns a channel that will receive changes when the underlying object is invalidated.
func (r *RemoteObject) ChangedChannel() chan struct{} {
	channel := make(chan struct{}, 16)
	r.changedChannels[channel] = true
	return channel
}

// RemoveChangeChannel unregisters a channel from further events.
func (r *RemoteObject) RemoveChangeChannel(channel chan struct{}) {
	r.mutex.Lock()
	if r.changedChannels[channel] != false {
		delete(r.changedChannels, channel)
		close(channel)
	}
	r.mutex.Unlock()
}

// Closed returns a channel that is closed when the remote object in Qlik Associative Engine is closed
func (r *RemoteObject) Closed() chan struct{} {
	return r.closedCh
}

func (r *RemoteObject) signalChanged() {
	r.mutex.Lock()
	for channel := range r.changedChannels {
		channel <- struct{}{}
	}
	r.mutex.Unlock()
}

func (r *RemoteObject) signalClosed() {
	r.mutex.Lock()
	close(r.closedCh)
	for channel := range r.changedChannels {
		if r.changedChannels[channel] != false {
			delete(r.changedChannels, channel)
			close(channel)
		}
	}
	// Clear it
	r.changedChannels = make(map[chan struct{}]bool)
	r.mutex.Unlock()
}

// Invokes a method on the remote object
func (r *RemoteObject) rpc(ctx context.Context, method string, apiResponse interface{}, params ...interface{}) error {
	invocationResponse := r.interceptorChain(ctx, &Invocation{RemoteObject: r, Method: method, Params: params})
	if invocationResponse.Error != nil {
    err := invocationResponse.Error
    if qe, ok := err.(*qixError); ok {
      qe.callParams = params
    }
		return err
	}
	if apiResponse != nil {
		err := json.Unmarshal(invocationResponse.Result, apiResponse)
		if err != nil {
			return err
		}
	}
	return nil
}

// newRemoteObject creates a new RemoteObject instance
func newRemoteObject(session *session, objectInterface *ObjectInterface) *RemoteObject {
	remoteObject := &RemoteObject{
		session:         session,
		ObjectInterface: objectInterface,
		changedChannels: make(map[chan struct{}]bool),
		mutex:           sync.Mutex{},
		closedCh:        make(chan struct{}),
	}
	// Signal that the object is by definition changed from the beginning
	return remoteObject
}

package enigma

import (
	"sync"
)

type (
	remoteObjectRegistry struct {
		mutex         sync.Mutex
		remoteObjects map[int]*RemoteObject
	}
)

func (r *remoteObjectRegistry) registerRemoteObject(rpcObject *RemoteObject) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.remoteObjects[rpcObject.Handle] = rpcObject
}

func (r *remoteObjectRegistry) unregisterRemoteObject(handle int) *RemoteObject {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	var rpcObject = r.remoteObjects[handle]
	delete(r.remoteObjects, handle)
	return rpcObject
}
func (r *remoteObjectRegistry) getOrCreateRemoteObject(session *session, objectInterface *ObjectInterface) *RemoteObject {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.remoteObjects[objectInterface.Handle] == nil {
		r.remoteObjects[objectInterface.Handle] = newRemoteObject(session, objectInterface)
	}
	return r.remoteObjects[objectInterface.Handle]
}

func (r *remoteObjectRegistry) getRemoteObject(handle int) *RemoteObject {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.remoteObjects[handle]
}

func (r *remoteObjectRegistry) handleUpdates(changed []int, closed []int) {
	changedObjects := make([]*RemoteObject, len(changed))
	closedObjects := make([]*RemoteObject, len(closed))
	r.mutex.Lock()
	for i, handle := range changed {
		changedObjects[i] = r.remoteObjects[handle]
	}
	for i, handle := range closed {
		closedObjects[i] = r.remoteObjects[handle]
		delete(r.remoteObjects, handle)
	}
	r.mutex.Unlock()

	// Signal outside of the mutex to avoid locking multiple locks simultaneously (deadlock risk)
	for _, x := range changedObjects {
		if x != nil {
			x.signalChanged()
		}
	}
	for _, x := range closedObjects {
		if x != nil {
			x.signalClosed()
		}
	}
}

func (r *remoteObjectRegistry) signalAllObjectsClosed() {
	r.mutex.Lock()
	oldRemoteObjects := r.remoteObjects
	r.remoteObjects = make(map[int]*RemoteObject)
	r.mutex.Unlock()

	// Signal outside of the mutex to avoid locking multiple locks simultaneously (deadlock risk)
	for _, remoteObject := range oldRemoteObjects {
		remoteObject.signalClosed()
	}
}

func newRemoteObjectRegistry() *remoteObjectRegistry {
	return &remoteObjectRegistry{remoteObjects: make(map[int]*RemoteObject)}
}

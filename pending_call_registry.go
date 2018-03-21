package enigma

import (
	"context"
	"sync"
	"time"
)

type (
	pendingCall struct {
		Response         *socketInput
		ID               int
		Done             chan error
		receiveTimestamp time.Time
		messageSize      int
	}

	pendingCallRegistry struct {
		callIDSeq     int
		mutex         sync.Mutex
		pendingCalls  map[int]*pendingCall
		terminalError error
	}

	reservedRequestIDKey struct{}
)

func (q *pendingCallRegistry) takeRequestID() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.callIDSeq++
	return q.callIDSeq
}

func (q *pendingCallRegistry) registerPendingCall(ctx context.Context) *pendingCall {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	reservedRequestID := ctx.Value(reservedRequestIDKey{})
	var id int
	if reservedRequestID != nil {
		id = reservedRequestID.(int)
	} else {
		q.callIDSeq++
		id = q.callIDSeq
	}
	pendingCall := &pendingCall{Done: make(chan error, 10), ID: id}
	q.pendingCalls[id] = pendingCall
	return pendingCall
}

func (q *pendingCallRegistry) removePendingCall(id int) *pendingCall {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	pendingCall := q.pendingCalls[id]
	delete(q.pendingCalls, id)
	return pendingCall
}

func newPendingCallRegistry() *pendingCallRegistry {
	return &pendingCallRegistry{callIDSeq: 0, pendingCalls: make(map[int]*pendingCall)}
}

// Creates a new context that contains a reserved JSON RPC protocol level request id.
// It can be for instance be useful when the request id used for upcoming call needs to be known.
func (q *pendingCallRegistry) WithReservedRequestID(ctx context.Context) (context.Context, int) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.callIDSeq++
	id := q.callIDSeq
	newContext := context.WithValue(ctx, reservedRequestIDKey{}, id)
	return newContext, id
}

func (q *pendingCallRegistry) closedWithError() error {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return q.terminalError
}

func (q *pendingCallRegistry) closeAllPendingCallsWithError(err error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	oldPendingCalls := q.pendingCalls
	q.pendingCalls = make(map[int]*pendingCall)

	q.terminalError = err
	for _, pendingCall := range oldPendingCalls {
		pendingCall.Done <- err
	}
}

func (q *pendingCallRegistry) pendingCallCount() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return len(q.pendingCalls)
}

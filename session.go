package enigma

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type (
	// Session handles the WebSocket connection to Qlik Associative Engine
	session struct {
		*pendingCallRegistry
		*remoteObjectRegistry
		*sessionMessages
		*sessionChangeLists
		socket                   Socket
		dialer                   *Dialer
		isOpen                   bool
		callIDSeq                int
		outgoingMessages         chan []byte
		disconnectedFromServerCh chan struct{}
		interceptorChain         InterceptorContinuation
	}

	// ChangeListsKey key for ChangeLists context value
	ChangeListsKey struct{}
	// ChangeLists list of changed and closed handles.
	ChangeLists struct {
		// Changed list of changed object handles or nil
		Changed []int
		// Closed  list of closed object handles or nil
		Closed []int
	}
)

func (q *session) connect(ctx context.Context, url string, httpHeader http.Header) error {
	// Connect websocket
	socket, err := q.dialer.CreateSocket(ctx, url, httpHeader)
	if err != nil {
		return err
	}
	q.socket = socket
	// Start reader/writer loops
	go q.mainSessionLoop()
	return nil
}

func (q *session) currentSocket() Socket {
	return q.socket
}

func convertToRawSliceIfNeeded(params []interface{}) interface{} {
	if params == nil {
		return []interface{}{}
	}
	if len(params) == 0 {
		return params
	}
	jsonRawSlice := make([]json.RawMessage, len(params))
	for i, x := range params {
		switch v := x.(type) {
		case json.RawMessage:
			jsonRawSlice[i] = v
		default:
			return params
		}
	}
	return jsonRawSlice
}

func (q *session) mainSessionLoop() {
	var socket = q.socket
	var wg sync.WaitGroup
	socketError := make(chan error, 5)

	if q.dialer.TrafficLogger != nil {
		q.dialer.TrafficLogger.Opened()
	}

	// Socket writer thread
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-socketError: //A socket error happened on the reader side
				return
			case outgoingMessage := <-q.outgoingMessages:
				err := socket.WriteMessage(1, outgoingMessage)
				if err != nil {
					socketError <- err
				}
			}
		}
	}()
	// Socket reader thread
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case err := <-socketError: //A socket error happened on the writer side
				q.closeAllPendingCallsWithError(err)
				q.signalAllObjectsClosed()
				return
			default:
				_, message, err := socket.ReadMessage()
				receiveTimestamp := time.Now()
				if err != nil {
					socketError <- err
					q.closeAllPendingCallsWithError(err)
					q.signalAllObjectsClosed()
					return
				}
				q.handleResponse(message, receiveTimestamp)
			}

		}
	}()
	wg.Wait()
	if q.dialer.TrafficLogger != nil {
		q.dialer.TrafficLogger.Closed()
	}
	q.socket.Close()
	q.closeAllSessionEventChannels()
	q.closeAllChangeListChannels()
	close(q.disconnectedFromServerCh)

}
func (q *session) handleResponse(message []byte, receiveTimestamp time.Time) {
	if q.dialer.TrafficLogger != nil {
		q.dialer.TrafficLogger.Received(message)
	}
	rpcResponse := &socketInput{}
	json.Unmarshal(message, rpcResponse)
	if rpcResponse.Method != "" { //This is a notification
		q.emitSessionMessage(rpcResponse.Method, rpcResponse.Params)
	} else {
		pendingCall := q.removePendingCall(rpcResponse.ID)
		q.emitChangeLists(rpcResponse.Change, rpcResponse.Close, pendingCall == nil) // Emit this before marking the pending call as done to make sure it is there when the pending call returns
		if pendingCall != nil {
			pendingCall.Response = rpcResponse
			pendingCall.receiveTimestamp = receiveTimestamp
			pendingCall.messageSize = len(message)
			pendingCall.Done <- nil
		}
	}
	q.handleUpdates(rpcResponse.Change, rpcResponse.Close)
}

func (q *session) sendCancelRequest(requestID int) {
	go func() {
		request := rpcInvocationRequest{Handle: -1, ID: q.takeRequestID(), Method: "CancelRequest", Params: []interface{}{requestID}}
		socketOutput := &socketOutput{rpcInvocationRequest: request, JSONRPC: "2.0"}
		message, _ := json.Marshal(socketOutput)

		q.outgoingMessages <- message
	}()
}

func (q *session) invokeRPC(ctx context.Context, invocation *Invocation) *InvocationResponse {
	invokeTimestamp := time.Now()

	// Change empty params to empty interface array
	params := invocation.Params
	if params == nil {
		params = []interface{}{}
	}

	if closedError := q.closedWithError(); closedError != nil {
		if metricsCollector := getMetricsCollector(ctx); metricsCollector != nil {
			// Store metrics if requested in the context
			metricsCollector.Lock()
			metricsCollector.metrics.InvocationRequestTimestamp = invokeTimestamp
			metricsCollector.metrics.SocketWriteTimestamp = invokeTimestamp
			metricsCollector.metrics.SocketReadTimestamp = invokeTimestamp
			metricsCollector.metrics.InvocationResponseTimestamp = invokeTimestamp
			metricsCollector.metrics.RequestMessageSize = 0
			metricsCollector.metrics.ResponseMessageSize = 0
			metricsCollector.Unlock()

		}
		return &InvocationResponse{Result: nil, RequestID: q.takeRequestID(), Error: closedError}
	}

	// Send message
	pendingCall := q.registerPendingCall(ctx)
	request := rpcInvocationRequest{Handle: invocation.RemoteObject.Handle, ID: pendingCall.ID, Method: invocation.Method, Params: params}
	socketOutput := &socketOutput{rpcInvocationRequest: request, JSONRPC: "2.0"}
	message, err := marshal(socketOutput)
	if err != nil {
		return &InvocationResponse{Result: nil, RequestID: pendingCall.ID, Error: err}
	}

	if q.dialer.TrafficLogger != nil {
		q.dialer.TrafficLogger.Sent(message)
	}

	sendTimestamp := time.Now()
	requestMessageSize := len(message)
	q.outgoingMessages <- message

	if metricsCollector := getMetricsCollector(ctx); metricsCollector != nil {
		defer func() {
			// Store metrics if requested in the context
			metricsCollector.Lock()
			metricsCollector.metrics.InvocationRequestTimestamp = invokeTimestamp
			metricsCollector.metrics.SocketWriteTimestamp = sendTimestamp
			metricsCollector.metrics.SocketReadTimestamp = pendingCall.receiveTimestamp
			metricsCollector.metrics.InvocationResponseTimestamp = time.Now()
			metricsCollector.metrics.RequestMessageSize = requestMessageSize
			metricsCollector.metrics.ResponseMessageSize = pendingCall.messageSize
			metricsCollector.Unlock()
		}()
	}
	select {
	case <-ctx.Done():
		q.removePendingCall(pendingCall.ID)
		q.sendCancelRequest(pendingCall.ID)
		return &InvocationResponse{Result: nil, RequestID: pendingCall.ID, Error: ctx.Err()}
	case err := <-pendingCall.Done:
		var result *InvocationResponse
		// In this case the pendingCall has already been removed from the pending call registry
		if err != nil {
			// Something bad happened during send/receive
			result = &InvocationResponse{Result: nil, RequestID: pendingCall.ID, Error: err}
		} else if pendingCall.Response.Error != nil {
			// error sent from server
			result = &InvocationResponse{Result: nil, RequestID: pendingCall.ID, Error: pendingCall.Response.Error}
		} else {
			result = &InvocationResponse{Result: *pendingCall.Response.Result, RequestID: pendingCall.ID, Error: nil}

			// Store change and close lists if requested in the context
			if cl := changeListFromContext(ctx); cl != nil {
				cl.Changed = pendingCall.Response.Change
				cl.Closed = pendingCall.Response.Close
			}
		}

		return result
	}
}

func (q *session) getRemoteObject(objectInterface *ObjectInterface) *RemoteObject {
	return q.getOrCreateRemoteObject(q, objectInterface)
}

func (q *session) GetMockSocket() *MockSocket {
	return q.socket.(*MockSocket)
}

func (q *session) DisconnectFromServer() {
	q.socket.Close()
	<-q.Disconnected()
}

// Disconnected returns a channel that will be closed once the underlying socket is closed.
func (q *session) Disconnected() chan struct{} {
	return q.disconnectedFromServerCh
}

// marshal works like json.Marshal but it customizes the encoder
// to not escape html.
func marshal(i interface{}) ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(i)
	b := buf.Bytes()
	// Without html-escaping we might get a trailing newline.
	if b[len(b)-1] == byte('\n') {
		b = b[:len(b)-1] // (In that case remove it!)
	}
	return buf.Bytes(), err
}

func newSession(dialer *Dialer) *session {
	qixSession := &session{
		socket:                   nil,
		dialer:                   dialer,
		sessionMessages:          newSessionEvents(),
		sessionChangeLists:       newSessionChangeLists(),
		outgoingMessages:         make(chan []byte, 50),
		pendingCallRegistry:      newPendingCallRegistry(),
		remoteObjectRegistry:     newRemoteObjectRegistry(),
		disconnectedFromServerCh: make(chan struct{}, 1),
	}
	qixSession.interceptorChain = buildInterceptorChain(dialer.Interceptors, qixSession.invokeRPC)

	return qixSession
}

func changeListFromContext(ctx context.Context) *ChangeLists {
	clk := ctx.Value(ChangeListsKey{})
	if clk != nil {
		if cl, ok := clk.(*ChangeLists); ok {
			return cl
		}
	}
	return nil
}

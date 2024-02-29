package enigma

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
)

type (
	// Invocation represents one invocation towards a remote object
	Invocation struct {
		// RemoteObject contains information about what Qlik Associative Engine object to call
		RemoteObject *RemoteObject
		// Method is the name of the method being called
		Method string
		// Params contains the function call parameters as provided in the top level API. Parameter types can be both primitives, structs and raw json (byte arrays) depending on what api level function is used.
		Params []interface{}
	}

	// InvocationResponse represents a QIX engine response message
	InvocationResponse struct {
		Result    json.RawMessage
		RequestID int
		Error     error
	}

	// Socket defines a set of functions that custom WebSocket implementations are expected to implement.
	Socket interface {
		WriteMessage(messageType int, data []byte) error
		ReadMessage() (int, []byte, error)
		Close() error
	}

	// TrafficLogger defines callback functions that can be used to log all network traffic
	TrafficLogger interface {
		Opened()
		Sent(message []byte)
		Received(message []byte)
		Closed()
	}

	// InterceptorContinuation executes the rest of the call chain. The call locks until the request is fulfilled and a response is returned
	InterceptorContinuation func(ctx context.Context, invocation *Invocation) *InvocationResponse

	// Interceptor is a function which takes an invocation request, forwards it to the next step in the interceptor chain.
	// It is synchronous and waits for the response from the rest of the call chain before returning the response.
	// This means that an interceptor can affect both the request and the response in one function call
	Interceptor func(ctx context.Context, invocation *Invocation, next InterceptorContinuation) *InvocationResponse

	// Dialer contains various settings for how to create WebSocket connections towards Qlik Associative Engine.
	Dialer struct {
		// A function to use when instantiating the WebSocket
		CreateSocket func(ctx context.Context, url string, httpHeader http.Header) (Socket, error)

		// A Config structure used to configure a TLS client or server
		TLSClientConfig *tls.Config

		// An array of interceptors that can be used to inject behaviour in the call chain
		Interceptors []Interceptor

		// An optional traffic logger. Note that this can not be used in conjunction with the TrafficDumpFile parameter.
		TrafficLogger TrafficLogger

		// Specifies the path to a protocol traffic log file. When the MockMode parameter is set to false the a traffic logger writes the traffic to the specified file.
		// If MockMode is set to true the requests and responses recorded in the log file are used to respond to QIX API calls - in effect replaying a previously recorded scenario.
		TrafficDumpFile string

		// When set to true a mock socket replaying previously recorded traffic is used instead of a real one.
		// TrafficDumpFile specified what log file to use.
		MockMode bool

		// Jar specifies the cookie jar.
		// If Jar is nil, cookies are not sent in requests and ignored
		// in responses.
		Jar http.CookieJar
	}
)

// DialRaw establishes a connection to Qlik Associative Engine using the settings set in the Dialer.
// The returned remote object points to the Global object of the session with handle -1.
// DialRaw can be used with custom specifications by wrapping the returned RemoteObject in a generated schema type like so:
//
//	remoteObject, err := enigma.Dialer{}.DialRaw(ctx, "ws://...", nil)
//	mySpecialGlobal := &specialSchemaPackage.SpecialGlobal{RemoteObject: remoteObject}
func (dialer Dialer) DialRaw(ctx context.Context, url string, httpHeader http.Header) (*RemoteObject, error) {
	// Set empty http header if omitted
	if httpHeader == nil {
		httpHeader = make(http.Header, 0)
	}

	if dialer.MockMode {
		dialer.CreateSocket = func(ctx context.Context, url string, httpHeader http.Header) (Socket, error) {
			socket, err := NewMockSocket(dialer.TrafficDumpFile)
			return socket, err
		}
	} else {
		// Create default CreateSocket function if omitted
		if dialer.CreateSocket == nil {
			setupDefaultDialer(&dialer)
		}
		if dialer.TrafficDumpFile != "" {
			dialer.TrafficLogger = newFileTrafficLogger(dialer.TrafficDumpFile)
		}
	}

	enigmaSession := newSession(&dialer)
	err := enigmaSession.connect(ctx, url, httpHeader)
	if err != nil {
		return nil, err
	}
	return enigmaSession.getRemoteObject(&ObjectInterface{Handle: -1, Type: "Global"}), nil
}

// Dial establishes a connection to Qlik Associative Engine using the settings set in the Dialer.
func (dialer Dialer) Dial(ctx context.Context, url string, httpHeader http.Header) (*Global, error) {
	remoteObject, err := dialer.DialRaw(ctx, url, httpHeader)
	if err != nil {
		return nil, err
	}
	return &Global{RemoteObject: remoteObject}, nil
}

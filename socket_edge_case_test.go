package enigma

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/websocket"
)

var serverAddr string
var once sync.Once
var originAndJwtHeaders = http.Header{"origin": []string{"http://localhost"}, "authorization": []string{"jwt content"}}

type Handler func(*websocket.Conn)

func checkOrigin(config *websocket.Config, req *http.Request) (err error) {
	config.Origin, err = websocket.Origin(config, req)
	if err == nil && config.Origin == nil {
		return fmt.Errorf("null origin")
	}
	return err
}

// ServeHTTP implements the http.Handler interface for a WebSocket
func (h Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Header.Get("authorization") == "" {
		w.WriteHeader(401)
		return
	}
	s := websocket.Server{Handler: func(c *websocket.Conn) {
		h(c)
	}, Handshake: checkOrigin}
	s.ServeHTTP(w, req)
}

func fakeEngineServer(waitTime time.Duration, error *qixError) Handler {
	return func(ws *websocket.Conn) {
		defer ws.Close()

		preMessages := []string{
			// Testing notifications
			`{"jsonrpc":"2.0","method":"OnConnected","params":{"qSessionState":"SESSION_CREATED"}}`,
			// Testing pending messages
			`{"jsonrpc":"2.0","id": 42}`,
		}

		// Send websocket messages that should be ignored
		for _, msg := range preMessages {
			err := websocket.Message.Send(ws, msg)
			if err != nil {
				return
			}
		}

		for {
			var req socketOutput
			err := websocket.JSON.Receive(ws, &req)
			if err != nil {
				return
			}

			var res interface{}
			if error != nil {
				res = socketInput{
					JSONRPC: req.JSONRPC,
					rpcInvocationResponse: rpcInvocationResponse{
						ID:    req.ID,
						Error: error,
					},
				}
			} else {
				result := json.RawMessage(`{ qHandle: 1, qType: "doc", qGenericID: "/apps/something" }`)
				res = socketInput{
					JSONRPC: req.JSONRPC,
					rpcInvocationResponse: rpcInvocationResponse{
						ID:     req.ID,
						Result: &result,
					},
				}
			}

			time.Sleep(waitTime)

			err = websocket.JSON.Send(ws, res)
			if err != nil {
				return
			}
		}
	}
}

func buildSuccessServer() Handler {
	return fakeEngineServer(0, nil)
}

func buildTimeoutServer() Handler {
	return fakeEngineServer(500*time.Millisecond, nil)
}

func buildMissingResultServer() Handler {
	return func(ws *websocket.Conn) {
		defer ws.Close()

		for {
			var req socketOutput
			err := websocket.JSON.Receive(ws, &req)
			if err != nil {
				return
			}

			err = websocket.Message.Send(ws, `{"jsonrpc":"2.0","id":1,"result":null}`)
			if err != nil {
				return
			}
		}
	}
}

type HandshakeTimeoutHandler struct{}

func (ct HandshakeTimeoutHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	time.Sleep(1000 * time.Millisecond)
	buildSuccessServer().ServeHTTP(w, req)
}

func buildErrorServer() Handler {
	return fakeEngineServer(0, &qixError{
		ErrorCode:      500,
		ErrorMessage:   "parameter",
		ErrorParameter: "error",
	})
}

func startServer() {
	http.Handle("/success", buildSuccessServer())
	http.Handle("/error", buildErrorServer())
	http.Handle("/missing-result", buildMissingResultServer())
	http.Handle("/timeout", buildTimeoutServer())
	http.Handle("/handshake-timeout", HandshakeTimeoutHandler{})
	http.Handle("/configureError", buildErrorServer())
	http.Handle("/doReloadError", buildErrorServer())
	http.Handle("/activeDocError", buildErrorServer())

	server := httptest.NewServer(nil)
	serverAddr = server.Listener.Addr().String()
}

func TestDial(t *testing.T) {
	once.Do(startServer)

	jwtHeader := http.Header{"authorization": []string{"jwt content"}}
	originHeader := http.Header{"origin": []string{"http://localhost"}}

	var dialTests = []struct {
		test             string
		url              string
		httpHeader       http.Header
		handshakeTimeout time.Duration
		expectedError    string
	}{
		{"success", "ws://" + serverAddr + "/success", originAndJwtHeaders, 0, ""},
		{"failure bad url", "//" + serverAddr + "/success", originHeader, 0, "malformed ws or wss URL"},
		{"failure bad origin", "ws://" + serverAddr + "/success", jwtHeader, 0, "403 from ws server"},
		{"failure no jwt", "ws://" + serverAddr + "/success", originHeader, 0, "401 from ws server"},
		{"failure with handshake timeout", "ws://" + serverAddr + "/handshake-timeout", originAndJwtHeaders, 5 * time.Millisecond, "context deadline exceeded"},
	}

	for _, tt := range dialTests {
		ctx := context.Background()
		if tt.handshakeTimeout != 0 {
			var cancel func()
			ctx, cancel = context.WithTimeout(ctx, tt.handshakeTimeout)
			defer cancel()
		}

		conn, err := Dialer{}.Dial(ctx, tt.url, tt.httpHeader)
		if tt.expectedError != "" {
			assert.Error(t, err, tt.test)
			assert.Contains(t, err.Error(), tt.expectedError, tt.test)
			continue
		} else {
			assert.NoError(t, err, tt.test)
		}
		conn.DisconnectFromServer()
	}
}

func TestOpenDocOnClosedConnectionError(t *testing.T) {
	once.Do(startServer)
	conn, err := Dialer{}.Dial(context.Background(), "ws://"+serverAddr+"/error", originAndJwtHeaders)
	assert.NoError(t, err)
	conn.DisconnectFromServer()

	doc, err := conn.OpenDoc(context.Background(), "appID", "", "", "", false)
	assert.Nil(t, doc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "use of closed network connection")
	assert.EqualValues(t, 0, conn.pendingCallCount())
}

func TestConnectionClosedDuringOpenDoc(t *testing.T) {
	once.Do(startServer)
	conn, err := Dialer{}.Dial(context.Background(), "ws://"+serverAddr+"/timeout", originAndJwtHeaders)
	assert.NoError(t, err)

	go func() {
		time.Sleep(1000 * time.Millisecond)
		conn.DisconnectFromServer()
	}()
	doc, err := conn.OpenDoc(context.Background(), "appID", "", "", "", false)
	fmt.Print(err)
	assert.Nil(t, doc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "websocket: close 1000 (normal)")
	assert.EqualValues(t, 0, conn.pendingCallCount())
}

func TestOpenDocSendTimeout(t *testing.T) {
	once.Do(startServer)

	conn, err := Dialer{}.Dial(context.Background(), "ws://"+serverAddr+"/timeout", originAndJwtHeaders)
	assert.NoError(t, err)
	defer conn.DisconnectFromServer()

	ctx, cancel := context.WithCancel(context.Background())
	// Cancelling before calling
	cancel()

	doc, err := conn.OpenDoc(ctx, "appID", "", "", "", false)
	assert.Nil(t, doc)
	assert.EqualError(t, err, "context canceled")
	assert.EqualValues(t, 0, conn.pendingCallCount())
}

func TestOpenDocTimeout(t *testing.T) {
	once.Do(startServer)

	conn, err := Dialer{}.Dial(context.Background(), "ws://"+serverAddr+"/timeout", originAndJwtHeaders)

	assert.EqualValues(t, 0, conn.pendingCallCount())
	assert.NoError(t, err)
	defer conn.DisconnectFromServer()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	doc, err := conn.OpenDoc(ctx, "appID", "", "", "", false)
	assert.Nil(t, doc)
	assert.EqualError(t, err, "context deadline exceeded")
}

func TestHTTPTrace(t *testing.T) {
	url := "ws://" + serverAddr + "/success"
	header := originAndJwtHeaders

	// See if the builtin http.ClientTrace can be provided in a
	// context to provide granular information about how the
	// HTTP request (the upgrade) was performed.
	c := 0
	ctx := httptrace.WithClientTrace(context.Background(), testTrace(&c))
	_, err := Dialer{}.Dial(ctx, url, header)
	if err != nil {
		t.Fatal("Connection error: ", err)
	}
	// If the counter hasn't been incremented to 15 then the
	// `http.ClientTrace` passed in the context wasn't correctly utilized.
	exp := 15
	if c != exp {
		t.Errorf("Expected counter to be %d but was %d", exp, c)
	}
}

// testTrace takes a counter. Each function called increments the counter
// by one.
func testTrace(counter *int) *httptrace.ClientTrace {
	return &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			fmt.Println("Get Connection:", hostPort)
			(*counter)++
		},
		GotConn: func(info httptrace.GotConnInfo) {
			fmt.Printf("Got Connection: %#v\n", info)
			(*counter)++
		},
		GotFirstResponseByte: func() {
			fmt.Println("First byte!")
			(*counter)++
		},
		Got100Continue: func() {
			fmt.Println("Got 100 continue")
			(*counter)++
		},
		DNSStart: func(info httptrace.DNSStartInfo) {
			fmt.Println("DNS Start: ", info.Host)
			(*counter)++
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			fmt.Printf("DNS Done: %#v\n", info)
			(*counter)++
		},
		ConnectStart: func(network, addr string) {
			fmt.Printf("Connect start: %s %s\n", network, addr)
			(*counter)++
		},
		ConnectDone: func(network, addr string, err error) {
			fmt.Printf("Connect done: %s %s - err: %v\n", network, addr, err)
			(*counter)++
		},
		TLSHandshakeStart: func() {
			fmt.Println("TLS Handshake Start")
			(*counter)++
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			var ver string
			switch state.Version {
			case tls.VersionTLS10:
				ver = "VersionTLS10"
			case tls.VersionTLS11:
				ver = "VersionTLS11"
			case tls.VersionTLS12:
				ver = "VersionTLS12"
			case tls.VersionTLS13:
				ver = "VersionTLS13"
			case tls.VersionSSL30:
				ver = "VersionTLS30"
			default:
				ver = "UNKNOWN!"
			}
			fmt.Printf("TLS Handshake (%s) Done: %#v - err: %v\n",
				ver, state, err)
			(*counter)++
		},
		WroteHeaderField: func(key string, value []string) {
			fmt.Printf("> %s: %s\n", key, strings.Join(value, ""))
			(*counter)++
		},
		WroteHeaders: func() {
			fmt.Println("Wrote Headers")
			(*counter)++
		},
		Wait100Continue: func() {
			fmt.Println("Waiting for 100 continue")
			(*counter)++
		},
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			fmt.Println("Wrote Request - err:", info.Err)
			(*counter)++
		},
	}
}

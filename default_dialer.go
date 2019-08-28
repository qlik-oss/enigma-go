package enigma

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"net"
	"net/http"
)

func setupDefaultDialer(dialer Dialer) {
	dialer.CreateSocket = func(ctx context.Context, url string, httpHeader http.Header) (Socket, error) {
		gorillaDialer := websocket.Dialer{
			TLSClientConfig: dialer.TLSClientConfig,
			NetDial: func(network, addr string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, network, addr)
			},
			Jar: dialer.Jar,
		}

		// Run the actual websocket dialing (including the upgrade) in a goroutine so we can
		// return if the context times out
		chConn := make(chan *websocket.Conn, 1)
		chErr := make(chan error, 1)
		go func() {
			conn, resp, err := gorillaDialer.Dial(url, httpHeader)
			if err == websocket.ErrBadHandshake {
				chErr <- errors.Wrapf(err, "%d from ws server", resp.StatusCode)
			} else if err != nil {
				chErr <- err
			} else {
				select {
				case <-ctx.Done():
					conn.Close()
				default:
					chConn <- conn
				}
			}
		}()
		select {
		case <-ctx.Done():
			return nil, errors.Wrapf(ctx.Err(), "error connecting to ws server %s", url)
		case err := <-chErr:
			return nil, err
		case ws := <-chConn:
			return ws, nil
		}
	}
}

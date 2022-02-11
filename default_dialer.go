package enigma

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

func setupDefaultDialer(dialer *Dialer) {
	dialer.CreateSocket = func(ctx context.Context, url string, httpHeader http.Header) (Socket, error) {
		gorillaDialer := websocket.Dialer{
			Proxy:           http.ProxyFromEnvironment, // Will pick the Proxy URL from the environment variables (HTTPS_PROXY).
			TLSClientConfig: dialer.TLSClientConfig,
			Jar:             dialer.Jar,
		}

		// Run the actual websocket dialing (including the upgrade) in a goroutine so we can
		// return if the context times out
		conn, resp, err := gorillaDialer.DialContext(ctx, url, httpHeader)
		if err != nil {
			if err == websocket.ErrBadHandshake {
				err = errors.Wrapf(err, "%d from ws server", resp.StatusCode)
			} else if strings.Contains(err.Error(), "Proxy Authentication") {
				err = fmt.Errorf("only proxies with http basic authentication are supported by enigma-go")
			} else if nerr, ok := err.(net.Error); ok {
				// For some reason the net package times out a tiny bit before
				// we get an error from ctx.Err(). To keep backwards compatability,
				// we want to return a "context deadline exceeded" error in this case
				// but it's not realiable to get it from ctx.Err. Therefore, we will
				// assume that a network timeout error is equivalent.
				if nerr.Timeout() {
					err = context.DeadlineExceeded
				}
			}
			return nil, err
		}
		return conn, nil
	}
}

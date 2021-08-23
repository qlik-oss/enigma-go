## PROXY

To use engima when working behind a proxy. Set the operating system environment variables called `HTTPS_PROXY`/`HTTP_PROXY` with the hostname or IP address of the proxy server.

The environment values may be either a complete URL or a "host[:port]", in which case the "http" scheme is assumed. The schemes "http", "https", and "socks5" are supported. An error is returned if the value is a different form.

https_proxy=https://\<ip-address>:\<port>  (or)  https_proxy=https://\<URL>

You must add the port number if the proxy server uses a port other than 80, as seen in the example below.
https_proxy=https://\<URL>:\<port>

**NOTE**
HTTPS_PROXY takes precedence over HTTP_PROXY for https requests.

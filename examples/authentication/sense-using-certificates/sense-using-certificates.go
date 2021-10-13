package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/qlik-oss/enigma-go/v3"
)

// the Qlik Associative Engine hostname:
const engineHost = "localhost"

// Make sure that the port below is accessible to the machine that runs this
// example. If you changed the Qlik Associative Engine port in your installation, change this:
const enginePort = 4747

// the user to impersonate when creating the session:
const userName = "<username>"

// the Qlik Sense Enterprise-configured user directory:
const userDirectory = "<user directory>"

// path to Sense Enterprise certificates:
const certificatesPath = "./certificates"

func main() {
	// Read client and root certificates.
	certFile := certificatesPath + "/client.pem"
	keyFile := certificatesPath + "/client_key.pem"
	caFile := certificatesPath + "/root.pem"

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		fmt.Println("Failed to load client certificate", err)
		panic(err)
	}

	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		fmt.Println("Failed to read root certificate", err)
		panic(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Setup TLS configuration.
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caCertPool,
	}

	ctx := context.Background()
	url := fmt.Sprintf("wss://%s:%d/app/engineData", engineHost, enginePort)

	// Notice how the user and directory is passed using the 'X-Qlik-User' header.
	headers := make(http.Header, 1)
	headers.Set("X-Qlik-User", fmt.Sprintf("UserDirectory=%s; UserId=%s", userDirectory, userName))

	global, err := enigma.Dialer{TLSClientConfig: tlsConfig}.Dial(ctx, url, headers)
	if err != nil {
		fmt.Println("Could not connect", err)
		panic(err)
	}
	version, _ := global.EngineVersion(ctx)
	global.DisconnectFromServer()
	fmt.Println(fmt.Sprintf("Connected to engine version %s.", version.ComponentVersion))
}

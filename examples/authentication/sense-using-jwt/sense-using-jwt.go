package main

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/qlik-oss/enigma-go/v2"
)

// the Qlik Sense Enterprise hostname:
const senseHost = "localhost"

// the virtual proxy configured for JWT authentication:
const proxyPrefix = "jwt"

// the user to impersonate when creating the session:
const userName = "<username>"

// the Qlik Sense Enterprise-configured user directory:
const userDirectory = "<user directory>"

// path to private key:
const privateKeyPath = "./keys/private.key"

// the JWT structure; change the attributes to match your virtual proxy configuration:
var jwtClaims = jwt.MapClaims{
	"user":      userName,
	"directory": userDirectory,
}

func main() {
	ctx := context.Background()
	key, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		fmt.Println("Could not find private key", err)
		panic(err)
	}
	decoded, _ := pem.Decode(key)
	privateKey, _ := x509.ParsePKCS1PrivateKey(decoded.Bytes)

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwtClaims)
	signedToken, _ := token.SignedString(privateKey)

	url := fmt.Sprintf("wss://%s/%s/app/engineData", senseHost, proxyPrefix)

	// Notice how the signed JWT is passed in the 'Authorization' header using the
	// 'Bearer' schema.
	headers := make(http.Header, 1)
	headers.Set("Authorization", fmt.Sprintf("Bearer %s", signedToken))

	global, err := enigma.Dialer{}.Dial(ctx, url, headers)
	if err != nil {
		fmt.Println("Could not connect", err)
		panic(err)
	}
	version, _ := global.EngineVersion(ctx)
	global.DisconnectFromServer()
	fmt.Println(fmt.Sprintf("Connected to engine version %s.", version.ComponentVersion))
}

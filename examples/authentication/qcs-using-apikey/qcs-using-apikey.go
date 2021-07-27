package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/qlik-oss/enigma-go/v2"
)

func main() {
	ctx := context.Background()

	tenant := "<tenant>"
	appId := "<appId>"
	url := fmt.Sprintf("wss://%s/app/%s", tenant, appId)

	qcsApiKey := "<qcsApiKey>"
	headers := make(http.Header, 1)
	headers.Set("Authorization", fmt.Sprintf("Bearer %s", qcsApiKey))

	global, err := enigma.Dialer{}.Dial(ctx, url, headers)
	if err != nil {
		fmt.Println("Could not connect", err)
		panic(err)
	}
	version, _ := global.EngineVersion(ctx)
	doc, err := global.GetActiveDoc(ctx)
	if err != nil {
		panic(err)
	}
	script, err := doc.GetScript(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println(script)

	global.DisconnectFromServer()
	fmt.Println(fmt.Sprintf("Connected to engine version %s.", version.ComponentVersion))
}

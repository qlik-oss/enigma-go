package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"

	"github.com/qlik-oss/enigma-go/v3"
)

type (
	// Augmented GenericObjectProperties with `meta` property
	CustomMetaProperties struct {
		enigma.GenericObjectProperties
		Meta `json:"meta"`
	}

	Meta struct {
		Title string `json:"title"`
	}
)

func main() {
	// Fetch the QCS_HOST and QCS_API_KEY from the environment variables
	QCS_HOST := os.Getenv("QCS_HOST")
	QCS_API_KEY := os.Getenv("QCS_API_KEY")

	// Connect to Qlik Cloud tenant and create a session document:
	ctx := context.Background()
	global, _ := enigma.Dialer{}.Dial(ctx, fmt.Sprintf("wss://%s/app/SessionApp_%v", QCS_HOST, rand.Int()), http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", QCS_API_KEY)},
	})

	doc, _ := global.GetActiveDoc(ctx)

	// Create 10 objects of type my-object with unique titles
	for i := 0; i < 10; i++ {
		properties := &CustomMetaProperties{
			GenericObjectProperties: enigma.GenericObjectProperties{
				Info: &enigma.NxInfo{
					Type: "my-object",
				},
			},
			Meta: Meta{
				Title: "my-object-" + strconv.Itoa(i),
			},
		}
		_, _ = doc.CreateObjectRaw(ctx, properties)

	}

	// Create a app object list using qAppObjectListDef and list all objects of type my-object
	// and also lists the title for each object.
	object, _ := doc.CreateObject(ctx, &enigma.GenericObjectProperties{
		Info: &enigma.NxInfo{
			Type: "my-list",
		},
		AppObjectListDef: &enigma.AppObjectListDef{
			Type: "my-object",
			Data: json.RawMessage(`{
				"title": "/meta/title"
			}`),
		},
	})

	layout, _ := object.GetLayout(ctx)

	LayoutAsJSON, _ := json.MarshalIndent(layout, "", "  ")
	fmt.Println(fmt.Sprintf("App object list: %s", LayoutAsJSON))

	// Close the session
	global.DisconnectFromServer()

}

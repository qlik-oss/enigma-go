package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/qlik-oss/enigma-go/v3"
)

const script = `
TempTable:
Load
RecNo() as Field1,
Rand() as Field2,
Rand() as Field3
AutoGenerate 100
`

func main() {

	// Open the session and create a session document:
	ctx := context.Background()
	global, _ := enigma.Dialer{}.Dial(ctx, "ws://localhost:9076/app/engineData", nil)

	doc, _ := global.CreateSessionApp(ctx)

	// Load in some data into the session document:
	doc.SetScript(ctx, script)
	doc.DoReload(ctx, 0, false, false)

	// Create a field list using qFieldListDef and list all fields available in the document.
	object, _ := doc.CreateObject(ctx, &enigma.GenericObjectProperties{
		Info: &enigma.NxInfo{
			Type: "my-field-list",
		},
		FieldListDef: &enigma.FieldListDef{},
	})

	layout, _ := object.GetLayout(ctx)

	LayoutAsJSON, _ := json.MarshalIndent(layout, "", "  ")
	fmt.Println(fmt.Sprintf("Field list: %s", LayoutAsJSON))

	// Close the session
	global.DisconnectFromServer()

}

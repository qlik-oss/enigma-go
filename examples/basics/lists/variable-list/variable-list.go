package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/qlik-oss/enigma-go"
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
	variable, _ := doc.CreateVariableEx(ctx, &enigma.GenericVariableProperties{
		Comment:    "sample comment",
		Definition: "=Count(Filed1)",
		Info: &enigma.NxInfo{
			Type: "variable",
		},
		Name: "vVariableName",
	})

	variable, _ = doc.GetVariableById(ctx, variable.GenericId)
	_, ok := interface{}(variable).(*enigma.GenericVariable)
	if !ok {
		fmt.Printf("GetVariableId returned wrong type: %T, should have been %T", variable, enigma.GenericVariable{})
		return
	}

	// Create a Variable list using qVariableListDef and list all variables available in the document.
	object, _ := doc.CreateSessionObject(ctx, &enigma.GenericObjectProperties{
		Info: &enigma.NxInfo{
			Type: "VariableList",
		},
		VariableListDef: &enigma.VariableListDef{
			Type: "variable",
			Data: json.RawMessage(`{
				"tags":"/tags"
			}`),
			ShowSession:  true,
			ShowConfig:   true,
			ShowReserved: true,
		},
	})

	layout, _ := object.GetLayoutRaw(ctx)

	LayoutAsJSON, _ := json.MarshalIndent(layout, "", "  ")
	fmt.Println(fmt.Sprintf("Variable list: %s", LayoutAsJSON))

	// Close the session
	global.DisconnectFromServer()

}

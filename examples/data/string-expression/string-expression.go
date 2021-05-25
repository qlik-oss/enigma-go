package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/qlik-oss/enigma-go/v2"
)

const script = `
TempTable:
Load
RecNo() as ID
AutoGenerate 100
`

type (
	// Augmented GenericObjectProperties with `expr` property
	CustomExpressionProperties struct {
		enigma.GenericObjectProperties
		Expr struct {
			StringExpression *enigma.StringExpression `json:"qStringExpression"`
		} `json:"expr"`
	}
	// Augmented GenericObjectLayout with `expr` property
	CustomExpressionLayout struct {
		*enigma.GenericObjectLayout
		Expr string `json:"expr"`
	}
)

func main() {

	ctx := context.Background()
	// Change to the path of your running Qlik Associative Engine.
	global, _ := enigma.Dialer{}.Dial(ctx, "ws://localhost:9076/app/engineData", nil)

	doc, _ := global.CreateSessionApp(ctx)
	doc.SetScript(ctx, script)
	doc.DoReload(ctx, 0, false, false)

	properties := &CustomExpressionProperties{
		GenericObjectProperties: enigma.GenericObjectProperties{
			Info: &enigma.NxInfo{
				Type: "StringExpression",
			},
		},
	}
	properties.Expr.StringExpression = &enigma.StringExpression{Expr: "='count(ID) = ' & count(ID)"}

	object, _ := doc.CreateObjectRaw(ctx, properties)

	layoutRaw, _ := object.GetLayoutRaw(ctx)

	customLayout := &CustomExpressionLayout{}
	json.Unmarshal(layoutRaw, customLayout)

	fmt.Println(fmt.Sprintf("Evaluated string expression: %s.", customLayout.Expr))

	global.DisconnectFromServer()
}

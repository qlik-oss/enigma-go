package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/qlik-oss/enigma-go/v4"
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
	// Fetch the QCS_HOST and QCS_API_KEY from the environment variables
	qcsHost := os.Getenv("QCS_HOST")
	qcsApiKey := os.Getenv("QCS_API_KEY")

	ctx := context.Background()
	rand.Seed(time.Now().UnixNano())
	// Connect to Qlik Cloud tenant and create a session document:
	global, _ := enigma.Dialer{}.Dial(ctx, fmt.Sprintf("wss://%s/app/SessionApp_%v", qcsHost, rand.Int()), http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", qcsApiKey)},
	})

	doc, _ := global.GetActiveDoc(ctx)
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

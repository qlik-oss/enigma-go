package enigma_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/qlik-oss/enigma-go"
	"github.com/stretchr/testify/assert"
	"testing"
)

//This example shows how to connect to a locally running Qlik Associative Engine, print the version number and disconnect again.
func Example() {
	ctx := context.Background()
	global, err := enigma.Dialer{}.Dial(ctx, "ws://localhost:9076", nil)
	if err != nil {
		panic(err)
	}
	engineVersion, err := global.EngineVersion(ctx)
	fmt.Println(engineVersion.ComponentVersion)
	global.DisconnectFromServer()
}

func TestFullRpcScenario(t *testing.T) {
	ctx := context.Background()

	global, _ := enigma.Dialer{MockMode: true}.Dial(context.Background(), "", nil)
	testSocket := global.GetMockSocket()

	testSocket.AddReceivedMessage(`{"jsonrpc":"2.0","method":"OnConnected","params":{"qSessionState":"SESSION_CREATED"}}`)
	testSocket.ExpectCall(
		`{"jsonrpc":"2.0","delta":false,"method":"OpenDoc","handle":-1,"id":1,"params":["doc","","","",false]}`,
		`{"jsonrpc":"2.0","id":1,"result":{"qReturn":{"qType":"Doc","qHandle":1,"qGenericId":"doc.qvf"}},"change":[1]}`)

	testSocket.ExpectCall(
		`{"jsonrpc":"2.0","delta":false,"method":"GetObject","handle":1,"id":2,"params":["hyperhyper"]}`,
		`{"jsonrpc":"2.0","id":2,"result":{"qReturn":{"qType":"GenericObject","qHandle":4,"qGenericType":"sheet","qGenericId":"JzJMza"}}}`)

	testSocket.ExpectCall(
		`{"jsonrpc":"2.0","delta":false,"method":"GetLayout","handle":4,"id":3,"params":[]}`,
		`{
		"jsonrpc": "2.0",
		"id": 3,
		"delta": false,
		"result": {
			"qLayout": {
				"qInfo": {
					"qId": "SheetList",
					"qType": "SheetList"
				},
				"qAppObjectList": {
					"qItems": [{
						"qInfo": {
							"qId": "GnAzpy",
							"qType": "sheet"
						},
						"qMeta": {
							"title": "Budget Analysis",
							"description": "Analyze actual versus budget budget data. Is the company on target to hit its budgeted amounts?"
						},
						"qData": {
							"customLayoutField": "customdata"
						}
					}]
				}
			}
		}
	}`)

	testSocket.ExpectCall(
		`{"jsonrpc":"2.0","delta":false,"method":"GetLayout","handle":4,"id":4,"params":[]}`,
		`{
			"jsonrpc": "2.0",
			"id": 4,
			"delta": false,
			"error": {
				"code": 123, 
				"parameter": "param",  
				"message":"mes"	
			}
		}`)

	// Check the OnConnected information
	sessionState, err := global.SessionState(ctx)

	assert.Equal(t, "SESSION_CREATED", sessionState)

	// Continue with opening the doc
	doc, _ := global.OpenDoc(ctx, "doc", "", "", "", false)
	obj, _ := doc.GetObject(ctx, "hyperhyper")

	type CustomLayout struct {
		Info          enigma.NxInfo `json:"qInfo,omitempty"`
		Meta          enigma.NxMeta `json:"qMeta,omitempty"`
		AppObjectList struct {
			Items []struct {
				enigma.NxContainerEntry
				Data struct {
					CustomLayoutField string `json:"customLayoutField,omitempty"`
				} `json:"qData,omitempty"`
			} `json:"qItems,omitempty"`
		} `json:"qAppObjectList,omitempty"`
	}

	layout := &CustomLayout{}
	layoutRaw, err := obj.GetLayoutRaw(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	json.Unmarshal(layoutRaw, layout)
	assert.Equal(t, layout.AppObjectList.Items[0].Info.Id, "GnAzpy")
	assert.Equal(t, layout.AppObjectList.Items[0].Data.CustomLayoutField, "customdata")

	_, err = obj.GetLayoutRaw(ctx)
	enigmaError := err.(enigma.Error)
	assert.Equal(t, enigmaError.Code(), 123)
	assert.Equal(t, enigmaError.Parameter(), "param")
	assert.Equal(t, enigmaError.Message(), "mes")

	testSocket.Close()
}

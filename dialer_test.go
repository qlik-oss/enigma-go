package enigma_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"testing"
	"time"

	"github.com/qlik-oss/enigma-go/v3"
	"github.com/stretchr/testify/assert"
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

func TestCookieJar(t *testing.T) {
	dialer := enigma.Dialer{MockMode: true}
	jar, err := cookiejar.New(nil)
	dialer.Jar = jar
	assert.NoError(t, err, "Error creating cookiejar")

	// Have some cookies!
	header := http.Header{}
	exp := fmt.Sprintf("%v", time.Now().Local().Add(time.Hour*time.Duration(48)).UTC())
	header.Add("Set-Cookie", "_session=a518840f-893b-4baf-bdf8-10d78ec14bf5; path=/; expires="+exp+"; secure; httponly")
	header.Add("Set-Cookie", "_grant=1d3cdfb9-25d0-42b2-8274-d4b11b97a475; path=/interaction/1d3cdfb9-25d0-42b2-8274-d4b11b97a475; expires="+exp+"; secure; httponly")
	header.Add("Set-Cookie", "_grant=1d3cdfb9-25d0-42b2-8274-d4b11b97a475; path=/auth/1d3cdfb9-25d0-42b2-8274-d4b11b97a475; expires="+exp+"; secure; httponly")

	response := http.Response{Header: header}
	cookies := response.Cookies()

	// Set the cookies
	url, err := url.Parse("https://www.qlik.com")
	assert.NoError(t, err, "Error parsing URL")
	dialer.Jar.SetCookies(url, cookies)

	// Test whether correct cookie is returned for request to /
	returnedCookies := dialer.Jar.Cookies(url)
	gotResponse := false
	for _, cookie := range returnedCookies {
		if cookie.Name == "_session" {
			assert.Equal(t, "a518840f-893b-4baf-bdf8-10d78ec14bf5", cookie.Value)
			gotResponse = true
		}
	}
	assert.True(t, gotResponse, fmt.Sprintf("Expected one cookie for path: <%v>", url.Path))

	// Test whether correct cookie is returns for request to interaction URL
	url, err = url.Parse("https://www.qlik.com/interaction/1d3cdfb9-25d0-42b2-8274-d4b11b97a475")
	returnedCookies = dialer.Jar.Cookies(url)
	assert.NoError(t, err, "Error parsing URL")
	gotResponse = false
	for _, cookie := range returnedCookies {
		if cookie.Name == "_grant" {
			assert.Equal(t, "1d3cdfb9-25d0-42b2-8274-d4b11b97a475", cookie.Value)
			gotResponse = true
		}
	}
	assert.True(t, gotResponse, fmt.Sprintf("Expected one cookie for path: <%v>", url.Path))
}

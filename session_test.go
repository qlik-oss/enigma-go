package enigma

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"net/http"
	"reflect"
	"testing"
)

func createAndConnectSession() (*session, *MockSocket, *RemoteObject) {
	// Setup qix session

	session := newSession(&Dialer{
		CreateSocket: func(ctx context.Context, url string, header http.Header) (Socket, error) { return NewMockSocket("") },
	})

	session.connect(context.Background(), "", nil)

	// Create rpc object
	rpcObject := session.getRemoteObject(&ObjectInterface{Handle: -1})

	return session, session.GetMockSocket(), rpcObject
}

func TestRpcInvocation(t *testing.T) {
	ctx := context.Background()

	_, testSocket, rpcObject := createAndConnectSession()

	testSocket.ExpectCall(
		`{"jsonrpc":"2.0","delta":false,"method":"DummyQixMethod","handle":-1,"id":1,"params":["a","b"]}`,
		`{"handle": -1, "id": 1, "result": "resultstring"}`)

	// Invoke rpc Method
	resultHolder := ""
	rpcerr := rpcObject.rpc(ctx, "DummyQixMethod", &resultHolder, "a", "b")

	// Expected response should be received
	assert.Equal(t, resultHolder, "resultstring")
	assert.Nil(t, rpcerr)

}

func TestRpcInvocationWithChangeListContext(t *testing.T) {
	cl := ChangeLists{}
	ctx := context.WithValue(context.Background(), ChangeListsKey{}, &cl)

	_, testSocket, rpcObject := createAndConnectSession()

	testSocket.ExpectCall(
		`{"jsonrpc":"2.0","delta":false,"method":"DummyQixMethod","handle":-1,"id":1,"params":["a","b"]}`,
		`{"handle": -1, "id": 1, "result": "resultstring","change":[1,7,8,9,10,11,12,13,14,15,16,17],"close":[5,6]}`)

	// Invoke rpc Method
	resultHolder := ""
	rpcerr := rpcObject.rpc(ctx, "DummyQixMethod", &resultHolder, "a", "b")

	// Expected response should be received
	assert.Equal(t, "resultstring", resultHolder)
	assert.Nil(t, rpcerr)

	expectedChanges := []int{1, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17}
	assert.Equal(t, expectedChanges, cl.Changed)

	expectedCloses := []int{5, 6}
	assert.Equal(t, expectedCloses, cl.Closed)

}

func TestClose(t *testing.T) {
	session, testSocket, _ := createAndConnectSession()
	// Close socket and wait for suspend event
	testSocket.Close()
	<-session.Disconnected()
}

func TestRawArrayConversion(t *testing.T) {
	p1 := json.RawMessage("string1")
	p2 := json.RawMessage("string2")
	params := []interface{}{p1, p2}
	result := convertToRawSliceIfNeeded(params)
	assert.Equal(t, "[]json.RawMessage", reflect.TypeOf(result).String())
}
func TestRawArrayConversionWithoutRawContent(t *testing.T) {
	p1 := "string1"
	p2 := "string2"
	params := []interface{}{p1, p2}
	result := convertToRawSliceIfNeeded(params)
	assert.Equal(t, params, result, "Should not convert the slice")
}

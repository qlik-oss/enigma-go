package enigma

import (
	"encoding/json"
)

type (
	rpcInvocationRequest struct {
		Method string      `json:"method"`
		Handle int         `json:"handle"`
		ID     int         `json:"id"`
		Params interface{} `json:"params"`
	}

	rpcInvocationResponse struct {
		ID     int              `json:"id"`
		Result *json.RawMessage `json:"result"`
		Error  *qixError        `json:"error"`
	}

	rpcNotification struct {
		Method string          `json:"method"`
		Params json.RawMessage `json:"params"`
	}

	onConnectedEvent struct {
		SessionState string `json:"qSessionState"`
	}

	rpcStatusInfo struct {
		Change  []int `json:"change"`
		Close   []int `json:"close"`
		Suspend []int `json:"suspend"`
	}

	// socketOutput represents a request message sent to Qlik Associative Engine
	socketOutput struct {
		JSONRPC string `json:"jsonrpc"`
		Delta   bool   `json:"delta"`
		rpcInvocationRequest
	}

	// socketInput is a response from an engine
	socketInput struct {
		JSONRPC               string `json:"jsonrpc"`
		Delta                 bool   `json:"delta"`
		rpcInvocationResponse        //Include fields for rpc request responses
		rpcNotification              //Include fields for notifications from the serverside
		rpcStatusInfo                //Include close/change fields
	}
)

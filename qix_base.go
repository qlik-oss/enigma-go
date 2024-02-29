package enigma

import (
	"encoding/json"
)

func ensureEncodable(val interface{}) interface{} {
	switch v := val.(type) {
	case []byte:
		return json.RawMessage(v)
	default:
		return val
	}
}

// Translates byte arrays into json.RawMessage
func ensureAllEncodable(params []interface{}) []interface{} {
	result := make([]interface{}, len(params), len(params))
	for i := range params {
		result[i] = ensureEncodable(params[i])
	}
	return result
}

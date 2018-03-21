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

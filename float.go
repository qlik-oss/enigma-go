package enigma

import (
	"encoding/json"
	"math"
)

//Float64 is an enigma-go equivalent of float64 which adds support for the Qlik Associative Engine specific way of marshalling and unmarshalling "Infinity", "-Infinity" and "NaN" as json strings.
//It can always safely be typecasted to plain float64 including these special cases.
type Float64 float64

// UnmarshalJSON implements the Unmarshaler interface for custom unmarshalling.
func (value *Float64) UnmarshalJSON(arg []byte) error {
	err := json.Unmarshal(arg, (*float64)(value))
	if err != nil {
		str := string(arg)
		switch str {
		case `"NaN"`:
			*value = Float64(math.NaN())
		case `"Infinity"`:
			*value = Float64(math.Inf(1))
		case `"-Infinity"`:
			*value = Float64(math.Inf(-1))
		default:
			return err
		}
	}
	return nil
}

// MarshalJSON implements the Marshaler interface for custom marshalling.
func (value Float64) MarshalJSON() ([]byte, error) {
	val := float64(value)
	if math.IsNaN(val) {
		return []byte(`"NaN"`), nil
	} else if math.IsInf(val, 1) {
		return []byte(`"Infinity"`), nil
	} else if math.IsInf(val, -1) {
		return []byte(`"-Infinity"`), nil
	}
	return json.Marshal(float64(value))
}

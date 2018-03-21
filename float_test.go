package enigma

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestFloat_UnmarshalJSON(t *testing.T) {
	var cell NxCell
	_ = json.Unmarshal([]byte(`{"qNum":"Infinity"}`), &cell)
	assert.True(t, math.IsInf(float64(cell.Num), 1))

	_ = json.Unmarshal([]byte(`{"qNum":"+Infinity"}`), &cell)
	assert.True(t, math.IsInf(float64(cell.Num), 1))

	_ = json.Unmarshal([]byte(`{"qNum":"-Infinity"}`), &cell)
	assert.True(t, math.IsInf(float64(cell.Num), -1))

	_ = json.Unmarshal([]byte(`{"qNum":"NaN"}`), &cell)
	assert.True(t, math.IsNaN(float64(cell.Num)))

	_ = json.Unmarshal([]byte(`{"qNum":-125}`), &cell)
	assert.Equal(t, float64(-125), float64(cell.Num))
}

func TestFloat_MarshalJSON(t *testing.T) {
	var cell NxCell
	var result []byte
	cell.Num = -125
	result, _ = json.Marshal(&cell)
	assert.Equal(t, `{"qNum":-125}`, string(result))

	cell.Num = Float64(math.Inf(+1))
	result, _ = json.Marshal(&cell)
	assert.Equal(t, `{"qNum":"Infinity"}`, string(result))

	cell.Num = Float64(math.Inf(-1))
	result, _ = json.Marshal(&cell)
	assert.Equal(t, `{"qNum":"-Infinity"}`, string(result))

	cell.Num = Float64(math.NaN())
	result, _ = json.Marshal(&cell)
	assert.Equal(t, `{"qNum":"NaN"}`, string(result))

}

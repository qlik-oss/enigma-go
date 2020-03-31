package enigma

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestFloatMarshalJSON(t *testing.T) {
	cases := []struct {
		float float64
		exp   string
	}{
		{math.NaN(), `"NaN"`},
		{math.Inf(1), `"Infinity"`},
		{math.Inf(-1), `"-Infinity"`},
		{1.0, "1"},
	}
	for _, cas := range cases {
		float := Float64(cas.float)
		ptr := &float
		var floatBytes, ptrBytes []byte
		var err error
		if floatBytes, err = json.Marshal(float); err != nil {
			t.Error(err)
			continue
		}
		if ptrBytes, err = json.Marshal(ptr); err != nil {
			t.Error(err)
			continue
		}
		if string(floatBytes) != string(ptrBytes) {
			t.Errorf("JSON should be equal for pointer and value")
			t.Log(" Float64:", string(floatBytes))
			t.Log("*Float64:", string(ptrBytes))
		}
		if string(floatBytes) != cas.exp {
			t.Errorf("Expected %q got %q", cas.exp, string(floatBytes))
		}
	}
}

func TestFloatUnmarshalJSON(t *testing.T) {
	cases := []struct {
		b   string
		exp Float64
	}{
		{`"NaN"`, Float64(math.NaN())},
		{`"Infinity"`, Float64(math.Inf(1))},
		{`"-Infinity"`, Float64(math.Inf(-1))},
		{"1", 1},
	}
	for _, cas := range cases {
		var f Float64
		err := json.Unmarshal([]byte(cas.b), &f)
		if err != nil {
			t.Error(err)
			continue
		}
		if f != cas.exp {
			if !(math.IsNaN(float64(f)) && math.IsNaN(float64(cas.exp))) {
				t.Errorf("Expected '%T: %v' got '%T: %v'",
					cas.exp, cas.exp, f, f)
			}
		}
	}
}

func TestFloatNonPointerUnmarshalJSON(t *testing.T) {
	var f Float64
	err := json.Unmarshal([]byte("1"), f)
	if err == nil {
		t.Error("Should not be able to unmarshal non-pointer")
	}
}

func TestFloatCompoundUnmarshalJSON(t *testing.T) {
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

func TestFloatCompoundMarshalJSON(t *testing.T) {
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

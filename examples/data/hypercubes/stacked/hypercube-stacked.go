package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/qlik-oss/enigma-go/v3"
)

const script = `
TempTable:
Load
RecNo() as ID,
RecNo()+1 as ID2,
Rand() as Value
AutoGenerate 100
`

func main() {
	// Fetch the QCS_HOST and QCS_API_KEY from the environment variables
	qcsHost := os.Getenv("QCS_HOST")
	qcsApiKey := os.Getenv("QCS_API_KEY")

	// Connect to Qlik Cloud tenant and create a session document:
	ctx := context.Background()
	rand.Seed(time.Now().UnixNano())
	global, _ := enigma.Dialer{}.Dial(ctx, fmt.Sprintf("wss://%s/app/SessionApp_%v", qcsHost, rand.Int()), http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", qcsApiKey)},
	})

	doc, _ := global.GetActiveDoc(ctx)
	doc.SetScript(ctx, script)
	doc.DoReload(ctx, 0, false, false)
	// Create a generic object with a hypercube stacked definition containing two dimensions and one measure
	object, _ := doc.CreateObject(ctx, &enigma.GenericObjectProperties{
		Info: &enigma.NxInfo{
			Type: "my-stacked-hypercube",
		},
		HyperCubeDef: &enigma.HyperCubeDef{
			Dimensions: []*enigma.NxDimension{{
				Def: &enigma.NxInlineDimensionDef{
					FieldDefs: []string{"ID"},
				},
			}, {
				Def: &enigma.NxInlineDimensionDef{
					FieldDefs: []string{"ID2"},
				},
			}},
			Measures: []*enigma.NxMeasure{{
				Def: &enigma.NxInlineMeasureDef{
					Def: "Sum(Value)",
				},
			}},
			Mode:                "EQ_DATA_MODE_PIVOT_STACK",
			AlwaysFullyExpanded: true,
		},
	})
	// Get hypercube stacked data
	data, _ := object.GetHyperCubeStackData(ctx, "/qHyperCubeDef", []*enigma.NxPage{{
		Top:    0,
		Left:   0,
		Height: 5,
		Width:  2,
	}}, 10000)

	HyperCubeDataPagesAsJSON, _ := json.MarshalIndent(data, "", "  ")

	fmt.Println(fmt.Sprintf("Hypercube data pages: %s", HyperCubeDataPagesAsJSON))
	// Select second value in the first column of the data matrix
	object.SelectPivotCells(ctx, "/qHyperCubeDef", []*enigma.NxSelectionCell{{
		Type: "D",
		Row:  1,
		Col:  0,
	}}, false, false)
	// Get stacked data
	data, _ = object.GetHyperCubeStackData(ctx, "/qHyperCubeDef", []*enigma.NxPage{{
		Top:    0,
		Left:   0,
		Height: 5,
		Width:  2,
	}}, 10000)

	HyperCubeDataPagesAsJSON, _ = json.MarshalIndent(data, "", "  ")

	fmt.Println(fmt.Sprintf("Hypercube data pages after selection: %s", HyperCubeDataPagesAsJSON))
	// Close the session
	global.DisconnectFromServer()

}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/qlik-oss/enigma-go/v4"
)

const script = `
TempTable:
Load
RecNo() as ID,
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
	// Create a generic object with a hypercube definition containing one dimension and one measure
	object, _ := doc.CreateObject(ctx, &enigma.GenericObjectProperties{
		Info: &enigma.NxInfo{
			Type: "my-straight-hypercube",
		},
		HyperCubeDef: &enigma.HyperCubeDef{
			Dimensions: []*enigma.NxDimension{{
				Def: &enigma.NxInlineDimensionDef{
					FieldDefs: []string{"ID"},
				},
			}},
			Measures: []*enigma.NxMeasure{{
				Def: &enigma.NxInlineMeasureDef{
					Def: "=Sum(Value)",
				},
			}},
			InitialDataFetch: []*enigma.NxPage{{
				Height: 5,
				Width:  2,
			}},
		},
	})

	// Get hypercube layout
	layout, _ := object.GetLayout(ctx)

	HyperCubeDataPagesAsJSON, _ := json.MarshalIndent(layout.HyperCube.DataPages, "", "  ")

	fmt.Println(fmt.Sprintf("Hypercube data pages: %s", HyperCubeDataPagesAsJSON))
	// Select cells at position 0, 2 and 4 in the dimension.
	object.SelectHyperCubeCells(ctx, "/qHyperCubeDef", []int{0, 2, 4}, []int{0}, false, false)
	// Get layout and view the selected values
	fmt.Println("After selection (notice the `qState` values)")
	layout, _ = object.GetLayout(ctx)

	HyperCubeDataPagesAsJSON, _ = json.MarshalIndent(layout.HyperCube.DataPages, "", "  ")
	fmt.Println(fmt.Sprintf("Hypercube data pages after selection: %s", HyperCubeDataPagesAsJSON))
	// Close the session
	global.DisconnectFromServer()

}

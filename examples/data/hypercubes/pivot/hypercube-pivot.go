package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/qlik-oss/enigma-go/v2"
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

	// Open the session and create a session document:
	ctx := context.Background()
	global, _ := enigma.Dialer{}.Dial(ctx, "ws://localhost:9076/app/engineData", nil)

	doc, _ := global.CreateSessionApp(ctx)
	doc.SetScript(ctx, script)
	doc.DoReload(ctx, 0, false, false)
	// Create a generic object with a hypercube pivot definition containing two dimensions and one measure
	object, _ := doc.CreateObject(ctx, &enigma.GenericObjectProperties{
		Info: &enigma.NxInfo{
			Type: "my-pivot-hypercube",
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
			Mode:                "EQ_DATA_MODE_PIVOT",
			AlwaysFullyExpanded: true,
		},
	})
	// Get hypercube pivot data
	data, _ := object.GetHyperCubePivotData(ctx, "/qHyperCubeDef", []*enigma.NxPage{{
		Top:    0,
		Left:   0,
		Height: 5,
		Width:  2,
	}})

	HyperCubeDataPagesAsJSON, _ := json.MarshalIndent(data, "", "  ")

	fmt.Println(fmt.Sprintf("Hypercube data pages: %s", HyperCubeDataPagesAsJSON))
	// Select second value in the first column of the data matrix
	object.SelectPivotCells(ctx, "/qHyperCubeDef", []*enigma.NxSelectionCell{{
		Type: "D",
		Row:  1,
		Col:  0,
	}}, false, false)
	// Get pivot data
	data, _ = object.GetHyperCubePivotData(ctx, "/qHyperCubeDef", []*enigma.NxPage{{
		Top:    0,
		Left:   0,
		Height: 5,
		Width:  2,
	}})

	HyperCubeDataPagesAsJSON, _ = json.MarshalIndent(data, "", "  ")

	fmt.Println(fmt.Sprintf("Hypercube data pages after selection: %s", HyperCubeDataPagesAsJSON))
	// Close the session
	global.DisconnectFromServer()

}

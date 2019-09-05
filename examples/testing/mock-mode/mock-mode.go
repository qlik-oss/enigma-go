package main

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"runtime"

	"github.com/qlik-oss/enigma-go"
)

const script = `
TempTable:
Load
RecNo() as ID,
Rand() as Value
AutoGenerate 100
`

func main() {
	_, filename, _, _ := runtime.Caller(0)
	trafficFileName := path.Dir(filename) + "/socket.traffic"

	liveDialer := &enigma.Dialer{TrafficDumpFile: trafficFileName, MockMode: false}
	mockDialer := &enigma.Dialer{TrafficDumpFile: trafficFileName, MockMode: true}

	fmt.Println("Running scenario against live Qlik Associative Engine while recording traffic")
	runScenarioWithDialer(liveDialer)

	fmt.Println("Running the same scenario against mock socket with replayed QIX traffic from the previous run")
	runScenarioWithDialer(mockDialer)
}

func runScenarioWithDialer(dialer *enigma.Dialer) {

	// Open the session and create a session document:
	ctx := context.Background()
	global, _ := dialer.Dial(ctx, "ws://localhost:9076/app/engineData", nil)
	doc, _ := global.CreateSessionApp(ctx)
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

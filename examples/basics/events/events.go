package main

import (
	"context"
	"fmt"
	"github.com/qlik-oss/enigma-go"
	"sync"
	"encoding/json"
	"time"
)

func main() {


	const script = "TempTable: Load RecNo() as ID, Rand() as Value AutoGenerate 1000000"
	ctx := context.Background()
	var waitGroup sync.WaitGroup
	waitGroup.Add(2)
	// Connect to Qlik Associative Engine.
	global, err := enigma.Dialer{}.Dial(ctx, "ws://localhost:9076", nil)
	if err != nil {
		fmt.Println("Could not connect", err)
		panic(err)
	}

	// Print messages coming in on the session
	sessionMessages := global.SessionMessageChannel()
	go func() {
		for sessionEvent := range sessionMessages {
			fmt.Println("Session message", sessionEvent.Topic, string(sessionEvent.Content))
		}
	}()
	// Once connected, create a session app and populate it with some data.
	doc, _ := global.CreateSessionApp(ctx)
	doc.SetScript(ctx, script)
	doc.DoReload(ctx, 0, false, false)
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

	layoutChangeChannel := object.ChangedChannel()
	layout, _ := object.GetLayout(ctx)

	HyperCubeDataPagesAsJSON, _ := json.MarshalIndent(layout.HyperCube.DataPages, "", "  ")
	fmt.Println(fmt.Sprintf("Initial hypercube layout: %s", HyperCubeDataPagesAsJSON))

	// Fetch additional layout changes in a separate goroutine by listening for change events
	go func() {
		for range layoutChangeChannel {
			layout, err = object.GetLayout(ctx)

			if err != nil {
				fmt.Println("The getlayout() call was aborted since the layout had already changed before we finished evaluating it")
				continue
			}
			HyperCubeDataPagesAsJSON, _ = json.MarshalIndent(layout.HyperCube.DataPages, "", "  ")
			fmt.Println(fmt.Sprintf("Changed hypercube layout: %s", HyperCubeDataPagesAsJSON))
		}
		fmt.Println("Layout change channel closed")
	}()
	// Change the selection. This will trigger a change event
	fmt.Println("Changing selection")
	object.SelectHyperCubeCells(ctx, "/qHyperCubeDef", []int{1, 3}, []int{0}, false, false)

	// To illustate the LOCERR_GENERIC_ABORTED error another selection is made shortly after the first one.
	// This means that a new selection is made in parallel with an ongoing getLayout call from the change event loop above.
	// Depending on timing this may or may not happen, but if it does, Qlik Associative Engine will return an error LOCERR_GENERIC_ABORTED
	// that means that the layout is already obsolete and a new one should be fetched.
	fmt.Println("Changing selection again")
	object.SelectHyperCubeCells(ctx, "/qHyperCubeDef", []int{0, 2, 4}, []int{0}, false, false)
	time.Sleep(1 * time.Second)


	// Destroy the object
	fmt.Println("Destroying object")
	doc.DestroyObject(ctx, object.GenericId)
	// Wait for it to be closed
	<- object.Closed()
	fmt.Println("Object closed")

	// Close the session
	time.Sleep(1 * time.Second)
	global.DisconnectFromServer()
}

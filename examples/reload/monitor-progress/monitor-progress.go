package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/qlik-oss/enigma-go"
	"time"
)

// Path to testdata, update to match your Qlik Associative Engine deployment
const testDataFolder = "/testdata"

func main() {
	ctx := context.Background()

	// Connect to Qlik Associative Engine
	global, err := enigma.Dialer{}.Dial(ctx, "ws://localhost:9076", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	// When we leave this function disconnect from Qlik Associative Engine
	defer global.DisconnectFromServer()

	// Open a session app that only lives in memory
	app, err := global.CreateSessionApp(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Create a folder connection that references the current directory where there is a test csv file to load.
	_, err = app.CreateConnection(ctx, &enigma.Connection{
		Type:             "folder",
		Name:             "testdata",
		ConnectionString: testDataFolder})
	if err != nil {
		fmt.Println(err)
		return
	}

	err = app.SetScript(ctx, `
		Airports:
		LOAD * FROM [lib://testdata/airports.csv]
		(txt, utf8, embedded labels, delimiter is ',', msq);`)
	if err != nil {
		fmt.Println(err)
		return
	}

	reloadDone := make(chan struct{})
	// The GetProgress call further down needs the protocol request id. So we reserve a request id that is used
	// in the DoReload request so we know what it will be.
	ctxWithReservedRequestID, reservedRequestID := app.WithReservedRequestID(ctx)
	// Monitor the reload
	go func() {
		for {
			select {
			case <-reloadDone:
				return
			default:
				time.Sleep(1000)
				// Get the progress using the request id we reserved for the reload
				progress, err := global.GetProgress(ctx, reservedRequestID)
				if err != nil {
					fmt.Println(err)
					return
				}
				progressBytes, _ := json.MarshalIndent(progress, "", "\t")
				fmt.Println(string(progressBytes))
			}

		}
	}()

	// Do the reload in a separate goroutine
	app.DoReload(ctxWithReservedRequestID, 0, false, false)
	close(reloadDone)
	fmt.Println("Reload Done")

	object, err := app.CreateSessionObject(ctx, &enigma.GenericObjectProperties{
		Info: &enigma.NxInfo{
			Type: "GenericObject",
			Id:   "AirportsExample",
		},
		HyperCubeDef: &enigma.HyperCubeDef{
			Dimensions: []*enigma.NxDimension{{
				Def: &enigma.NxInlineDimensionDef{
					FieldDefs: []string{
						"Airport",
					},
					SortCriterias: []*enigma.SortCriteria{{
						SortByAscii: 1,
					}},
				}},
			},
			InitialDataFetch: []*enigma.NxPage{{0, 0, 10, 10}},
		},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	layout, err := object.GetLayout(ctx)
	cell := layout.HyperCube.DataPages[0].Matrix[3][0]
	fmt.Println(cell.Text)
	// Output: 7 Novembre
}

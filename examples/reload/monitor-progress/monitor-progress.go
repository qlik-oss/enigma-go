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

func main() {
	// Fetch the QCS_HOST and QCS_API_KEY from the environment variables
	qcsHost := os.Getenv("QCS_HOST")
	qcsApiKey := os.Getenv("QCS_API_KEY")

	ctx := context.Background()
	rand.Seed(time.Now().UnixNano())

	// Connect to Qlik Cloud tenant and create a session document:
	global, err := enigma.Dialer{}.Dial(ctx, fmt.Sprintf("wss://%s/app/SessionApp_%v", qcsHost, rand.Int()), http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", qcsApiKey)},
	})
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	// When we leave this function disconnect from Qlik Associative Engine
	defer global.DisconnectFromServer()

	// Open a session app that only lives in memory
	app, err := global.GetActiveDoc(ctx)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	err = app.SetScript(ctx, `
	Characters:
	Load Chr(RecNo()+Ord('A')-1) as Alpha, RecNo() as Num autogenerate 26;
	 
	ASCII:
	Load 
	 if(RecNo()>=65 and RecNo()<=90,RecNo()-64) as Num,
	 Chr(RecNo()) as AsciiAlpha, 
	 RecNo() as AsciiNum
	autogenerate 255
	 Where (RecNo()>=32 and RecNo()<=126) or RecNo()>=160 ;
	 
	Transactions:
	Load
	 TransLineID, 
	 TransID,
	 mod(TransID,26)+1 as Num,
	 Pick(Ceil(3*Rand1),'A','B','C') as Dim1,
	 Pick(Ceil(6*Rand1),'a','b','c','d','e','f') as Dim2,
	 Pick(Ceil(3*Rand()),'X','Y','Z') as Dim3,
	 Round(1000*Rand()*Rand()*Rand1) as Expression1,
	 Round(  10*Rand()*Rand()*Rand1) as Expression2,
	 Round(Rand()*Rand1,0.00001) as Expression3;
	Load 
	 Rand() as Rand1,
	 IterNo() as TransLineID,
	 RecNo() as TransID
	Autogenerate 1000
	 While Rand()<=0.5 or IterNo()=1;`)
	if err != nil {
		fmt.Println(err)
		panic(err)
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
	reloadSuccessful, err := app.DoReload(ctxWithReservedRequestID, 0, false, false)

	if err != nil {
		fmt.Println("Error when reloading app", err)
		panic(err)
	}

	if !reloadSuccessful {
		panic("DoReload was not successful!")
	}

	close(reloadDone)
	fmt.Println("Reload Done")

	object, err := app.CreateSessionObject(ctx, &enigma.GenericObjectProperties{
		Info: &enigma.NxInfo{
			Type: "GenericObject",
			Id:   "TransactionsExample",
		},
		HyperCubeDef: &enigma.HyperCubeDef{
			Dimensions: []*enigma.NxDimension{{
				Def: &enigma.NxInlineDimensionDef{
					FieldDefs: []string{
						"Dim1",
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
		panic(err)
	}
	layout, err := object.GetLayout(ctx)
	cell := layout.HyperCube.DataPages[0].Matrix[1][0]
	fmt.Println(cell.Text)
	// Output: B
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"

	"github.com/qlik-oss/enigma-go/v3"
)

func main() {
	// Fetch the QCS_HOST and QCS_API_KEY from the environment variables
	QCS_HOST := os.Getenv("QCS_HOST")
	QCS_API_KEY := os.Getenv("QCS_API_KEY")

	const script = "TempTable: Load RecNo() as ID, Rand() as Value AutoGenerate 100"
	listObjectProperties := enigma.GenericObjectProperties{
		Info: &enigma.NxInfo{
			Type: "my-list-object",
		},
		ListObjectDef: &enigma.ListObjectDef{
			Def: &enigma.NxInlineDimensionDef{
				FieldDefs: []string{"Value"},
				SortCriterias: []*enigma.SortCriteria{
					{SortByLoadOrder: 1},
				},
			},
			ShowAlternatives: true,
			InitialDataFetch: []*enigma.NxPage{
				{
					Top:    0,
					Height: 3,
					Left:   0,
					Width:  1,
				},
			},
		},
	}

	ctx := context.Background()
	// Connect to Qlik Cloud tenant and create a session document:
	global, err := enigma.Dialer{}.Dial(ctx, fmt.Sprintf("wss://%s/app/SessionApp_%v", QCS_HOST, rand.Int()), http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", QCS_API_KEY)},
	})

	if err != nil {
		fmt.Println("Not able to connect", err)
		panic(err)
	}

	doc, _ := global.GetActiveDoc(ctx)
	doc.SetScript(ctx, script)
	doc.DoReload(ctx, 0, false, false)
	listObject, _ := doc.CreateObject(ctx, &listObjectProperties)
	fmt.Println("### No selection;")
	layout, _ := listObject.GetLayout(ctx)
	printLayoutInfo(layout)

	listObject.SelectListObjectValues(ctx, "/qListObjectDef", []int{0}, false, false)
	fmt.Println("### After selection (notice the `qState` values):")
	layout, _ = listObject.GetLayout(ctx)
	printLayoutInfo(layout)
	global.DisconnectFromServer()
}

func printLayoutInfo(layout *enigma.GenericObjectLayout) {
	fmt.Printf("Generic object info:\n%s\n", toString(layout.Info))
	fmt.Printf("List object state:\n%s\n", toString(layout.ListObject.DimensionInfo.StateCounts))
	fmt.Printf("List object data:\n%s\n", toString(layout.ListObject.DataPages[0].Matrix))
}

func toString(object interface{}) string {
	data, _ := json.MarshalIndent(object, "", "  ")
	return string(data)
}

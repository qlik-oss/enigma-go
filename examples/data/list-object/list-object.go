package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/qlik-oss/enigma-go"
)

func main() {
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

	global, err := enigma.Dialer{}.Dial(ctx, "ws://localhost:9076", nil)

	if err != nil {
		fmt.Println("Not able to connect", err)
		panic(err)
	}

	doc, _ := global.CreateSessionApp(ctx)
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

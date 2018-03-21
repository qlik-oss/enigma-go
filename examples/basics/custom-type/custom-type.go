package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/qlik-oss/enigma-go"
)

type (
	// Define custom aware Doc and GenericObject wrapper structs
	customDoc struct {
		*enigma.Doc
	}
	customObject struct {
		*enigma.GenericObject
	}

	// Define custom object properties, extending GenericObjectProperties
	customObjectProperties struct {
		enigma.GenericObjectProperties
		CustomProperty string `json:"customProperty"`
	}
)

func main() {
	// Connect to a local Qlik Associative Engine
	ctx := context.Background()
	global, err := enigma.Dialer{}.Dial(ctx, "ws://localhost:9076", nil)

	if err != nil {
		fmt.Println("Not able to connect", err)
		panic(err)
	}

	// Create a session app and cast it to custom object aware type
	d, _ := global.CreateSessionApp(ctx)
	doc := customDoc{d}


	// Create a custom object
	properties := &customObjectProperties{
		GenericObjectProperties: enigma.GenericObjectProperties{
			Info: &enigma.NxInfo{
				Type: "custom-object",
			},
		},
	}
	object, _ := doc.CreateCustomObject(ctx, properties)

	// Update a custom property
	properties.CustomProperty = "custom-property-value"
	_ = object.SetPropertiesRaw(ctx, properties)

	// Read properties back and print the custom property value
	fetchedProperties, _ := object.GetCustomProperties(ctx)
	fmt.Println(fmt.Sprintf("CustomProperty value is: %s", fetchedProperties.CustomProperty))

	global.DisconnectFromServer()
}

func (c *customDoc) CreateCustomObject(ctx context.Context, properties *customObjectProperties) (*customObject, error) {
	obj, err := c.CreateObjectRaw(ctx, properties)
	if err != nil {
		return nil, err
	}
	return &customObject{obj}, nil
}

func (c *customObject) GetCustomProperties(ctx context.Context) (*customObjectProperties, error) {
	rawProperties, err := c.GetPropertiesRaw(ctx)
	properties := &customObjectProperties{}
	err = json.Unmarshal(rawProperties, properties)
	if err != nil {
		return nil, err
	}
	return properties, nil
}

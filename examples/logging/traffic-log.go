package main

import (
	"context"
	"fmt"

	"github.com/qlik-oss/enigma-go"
)

const script = `
TempTable:
Load
RecNo() as Field1,
Rand() as Field2,
Rand() as Field3
AutoGenerate 100
`

// Implement the TrafficLogger interface
type Logger struct{}

func (l *Logger) Opened() {
	fmt.Println("Logger opened")
}

func (l *Logger) Closed() {
	fmt.Println("Logger closed")
}

func (l *Logger) Sent(message []byte) {
	fmt.Println("Sent:", string(message[:]))
}

func (l *Logger) Received(message []byte) {
	fmt.Println("Received:", string(message[:]))
}

func main() {

	// Define a logger and pass it to the Dialer
	var logger = &Logger{}
	ctx := context.Background()
	dialer := &enigma.Dialer{TrafficLogger: logger}
	global, _ := dialer.Dial(ctx, "ws://localhost:9076/app/engineData", nil)

	// Create a session app, set a script and perform a reload
	doc, _ := global.CreateSessionApp(ctx)

	// Load in some data into the session document:
	doc.SetScript(ctx, script)
	doc.DoReload(ctx, 0, false, false)

	// Close the session
	global.DisconnectFromServer()

}

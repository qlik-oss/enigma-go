package main

import (
	"context"
	"fmt"
	"path"
	"runtime"

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

	// Log JSON traffic to a file
	_, filename, _, _ := runtime.Caller(0)
	trafficFileName := path.Dir(filename) + "/socket.traffic"
	logToFileDialer := &enigma.Dialer{TrafficDumpFile: trafficFileName}
	runScenario(logToFileDialer)

	// Log JSON traffic to stdout
	var logger = &Logger{}
	logStdOutDialer := &enigma.Dialer{TrafficLogger: logger}
	runScenario(logStdOutDialer)
}

func runScenario(dialer *enigma.Dialer) {
	ctx := context.Background()
	global, _ := dialer.Dial(ctx, "ws://localhost:9076/app/engineData", nil)

	// Create a session app, set a script and perform a reload
	doc, _ := global.CreateSessionApp(ctx)
	doc.SetScript(ctx, script)
	doc.DoReload(ctx, 0, false, false)

	// Close the session
	global.DisconnectFromServer()
}

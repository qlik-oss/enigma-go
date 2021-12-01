package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path"
	"runtime"

	"github.com/qlik-oss/enigma-go/v3"
)

const script = `
TempTable:
Load
RecNo() as Field1,
Rand() as Field2,
Rand() as Field3
AutoGenerate 100
`

// Logger implements the TrafficLogger interface
type Logger struct{}

// Opened implements the Opened method in the TrafficLogger interface
func (Logger) Opened() {
	fmt.Println("Logger opened")
}

// Closed implements the Closed method in the TrafficLogger interface
func (Logger) Closed() {
	fmt.Println("Logger closed")
}

// Sent implements the Sent method in the TrafficLogger interface
func (Logger) Sent(message []byte) {
	fmt.Println("Sent:", string(message))
}

// Received implements the Received method in the TrafficLogger interface
func (Logger) Received(message []byte) {
	fmt.Println("Received:", string(message))
}

func main() {
	// Log JSON traffic to a file
	_, filename, _, _ := runtime.Caller(0)
	trafficFileName := path.Dir(filename) + "/socket.traffic"
	logToFileDialer := &enigma.Dialer{TrafficDumpFile: trafficFileName}
	runScenario(logToFileDialer)

	// Log JSON traffic to stdout
	logStdOutDialer := &enigma.Dialer{TrafficLogger: &Logger{}}
	runScenario(logStdOutDialer)
}

func runScenario(dialer *enigma.Dialer) {
	// Fetch the QCS_HOST and QCS_API_KEY from the environment variables
	QCS_HOST := os.Getenv("QCS_HOST")
	QCS_API_KEY := os.Getenv("QCS_API_KEY")
	ctx := context.Background()
	// Connect to Qlik Cloud tenant and create a session document:
	global, _ := dialer.Dial(ctx, fmt.Sprintf("wss://%s/app/SessionApp_%v", QCS_HOST, rand.Int()), http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", QCS_API_KEY)},
	})

	// Create a session app, set a script and perform a reload
	doc, _ := global.GetActiveDoc(ctx)
	doc.SetScript(ctx, script)
	doc.DoReload(ctx, 0, false, false)

	// Close the session
	global.DisconnectFromServer()
}

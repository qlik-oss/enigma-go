package main

import (
	"context"
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

	const script = "TempTable: Load RecNo() as ID, Rand() as Value AutoGenerate 1000000"
	ctx := context.Background()
	rand.Seed(time.Now().UnixNano())

	// Configure the dialer to use an interceptor.
	dialer := enigma.Dialer{
		Interceptors: []enigma.Interceptor{
			metricsInterceptor,
		},
	}

	// Connect to Qlik Cloud tenant and create a session document:
	global, err := dialer.Dial(ctx, fmt.Sprintf("wss://%s/app/SessionApp_%v", qcsHost, rand.Int()), http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", qcsApiKey)},
	})

	if err != nil {
		fmt.Println("Could not connect", err)
		panic(err)
	}

	// Once connected, create a session app and populate it with some data.
	doc, _ := global.GetActiveDoc(ctx)
	doc.SetScript(ctx, script)
	doc.DoReload(ctx, 0, false, false)

	global.DisconnectFromServer()
}

// MetricsInterceptor shows how to an interceptor that collects metrics data can be written.
func metricsInterceptor(ctx context.Context, invocation *enigma.Invocation, proceed enigma.InterceptorContinuation) *enigma.InvocationResponse {
	ctxWithMetrics, metricsCollector := enigma.WithMetricsCollector(ctx)

	response := proceed(ctxWithMetrics, invocation)

	metrics := metricsCollector.Metrics()
	fmt.Println(response.RequestID, invocation.RemoteObject.Type, invocation.Method, metrics.ToString())
	return response
}

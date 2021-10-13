package main

import (
	"context"
	"fmt"

	"github.com/qlik-oss/enigma-go/v3"
)

func main() {
	const script = "TempTable: Load RecNo() as ID, Rand() as Value AutoGenerate 1000000"
	ctx := context.Background()

	// Configure the dialer to use an interceptor.
	dialer := enigma.Dialer{
		Interceptors: []enigma.Interceptor{
			metricsInterceptor,
		},
	}

	// Connect to Qlik Associative Engine
	global, err := dialer.Dial(ctx, "ws://localhost:9076", nil)

	if err != nil {
		fmt.Println("Could not connect", err)
		panic(err)
	}

	// Once connected, create a session app and populate it with some data.
	doc, _ := global.CreateSessionApp(ctx)
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

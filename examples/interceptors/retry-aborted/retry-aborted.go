package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/qlik-oss/enigma-go/v4"
)

const LOCERR_GENERIC_ABORTED = 15
const MAX_RETRIES = 3

func main() {
	// Fetch the QCS_HOST and QCS_API_KEY from the environment variables
	qcsHost := os.Getenv("QCS_HOST")
	qcsApiKey := os.Getenv("QCS_API_KEY")

	const script = "TempTable: Load RecNo() as ID, Rand() as Value AutoGenerate 1000000"
	ctx := context.Background()
	rand.Seed(time.Now().UnixNano())
	var waitGroup sync.WaitGroup
	waitGroup.Add(2)

	// Configure the dialer to use an interceptor.
	dialer := enigma.Dialer{
		Interceptors: []enigma.Interceptor{
			retryAborted,
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

	// Start two goroutines: One that repeatedly invalidates the data model, and one
	// that evaluates an expression. The invalidation will cause evaluation to be aborted.
	go invalidate(ctx, &waitGroup, doc)
	go evaluate(ctx, &waitGroup, doc)

	waitGroup.Wait()
	global.DisconnectFromServer()
}

func retryAborted(ctx context.Context, invocation *enigma.Invocation, next enigma.InterceptorContinuation) *enigma.InvocationResponse {
	var response *enigma.InvocationResponse
	var retries int
	for {
		response = next(ctx, invocation)
		// Check the error to see if the call was aborted and should be retried.
		if qixErr, ok := response.Error.(enigma.Error); ok && qixErr.Code() == LOCERR_GENERIC_ABORTED && retries < MAX_RETRIES {
			retries++
			fmt.Println(fmt.Sprintf("Call to %s was aborted, retrying... (attempt %d of %d)", invocation.Method, retries, MAX_RETRIES))
			continue
		}
		break
	}
	return response
}

func invalidate(ctx context.Context, waitGroup *sync.WaitGroup, doc *enigma.Doc) {
	defer waitGroup.Done()
	for i := 0; i < 3; i++ {
		doc.ClearAll(ctx, false, "")
	}
}

func evaluate(ctx context.Context, waitGroup *sync.WaitGroup, doc *enigma.Doc) {
	defer waitGroup.Done()
	count, err := doc.Evaluate(ctx, "COUNT(Value)")
	if err == nil {
		fmt.Println(fmt.Sprintf("Evaluation completed, result: %s", count))
	} else {
		fmt.Println(fmt.Sprintf("Evaluation failed, error is: %s", err.Error()))
	}
}

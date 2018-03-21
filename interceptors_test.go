package enigma

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuildInterceptorChain(t *testing.T) {

	log := ""

	finalFunc := func(ctx context.Context, invocation *Invocation) *InvocationResponse {
		log += "<jsonrpc>"
		return nil
	}

	interceptor1 := func(ctx context.Context, invocation *Invocation, next InterceptorContinuation) *InvocationResponse {
		log += "<before1>"
		response := next(ctx, invocation)
		log += "<after1>"
		return response
	}

	interceptor2 := func(ctx context.Context, invocation *Invocation, next InterceptorContinuation) *InvocationResponse {
		log += "<before2>"
		response := next(ctx, invocation)
		log += "<after2>"
		return response
	}

	interceptorChain0 := buildInterceptorChain([]Interceptor{}, finalFunc)
	interceptorChain1 := buildInterceptorChain([]Interceptor{interceptor1}, finalFunc)
	interceptorChain2 := buildInterceptorChain([]Interceptor{interceptor1, interceptor2}, finalFunc)

	log = ""
	interceptorChain0(nil, &Invocation{RemoteObject: nil, Method: "", Params: nil})
	assert.Equal(t, "<jsonrpc>", log)

	log = ""
	interceptorChain1(nil, &Invocation{RemoteObject: nil, Method: "", Params: nil})
	assert.Equal(t, "<before1><jsonrpc><after1>", log)

	log = ""
	interceptorChain2(nil, &Invocation{RemoteObject: nil, Method: "", Params: nil})
	assert.Equal(t, "<before1><before2><jsonrpc><after2><after1>", log)

}

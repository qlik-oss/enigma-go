package enigma

import "context"

func buildInterceptorChain(interceptors []Interceptor, finalContinuation InterceptorContinuation) InterceptorContinuation {
	nextInvoker := finalContinuation
	for i := len(interceptors) - 1; i >= 0; i-- {
		nextInvoker = createContinuationFunction(interceptors[i], nextInvoker)
	}
	return nextInvoker
}

func createContinuationFunction(interceptor Interceptor, nextInvoker InterceptorContinuation) InterceptorContinuation {
	interceptorInvokerFunction := func(ctx context.Context, invocation *Invocation) *InvocationResponse {
		return interceptor(ctx, invocation, nextInvoker)
	}
	return interceptorInvokerFunction
}

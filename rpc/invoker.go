package rpc

import (
	"context"
)

type CallbackFunc func(*Response)

func (c CallbackFunc) Error(err Error) {
	c(&Response{Error: err})
}

type Interceptor interface {
	Invoke(ctx context.Context, req *Request, callback CallbackFunc, next Invoker)
}

type InterceptorFunc func(ctx context.Context, req *Request, callback CallbackFunc, next Invoker)

func (fn InterceptorFunc) Invoke(ctx context.Context, req *Request, callback CallbackFunc, next Invoker) {
	fn(ctx, req, callback, next)
}

type Invoker interface {
	Invoke(ctx context.Context, req *Request, callback CallbackFunc)
}

type ServiceInvoker interface {
	ServiceInfo() *ServiceInfo
	Invoker
}

type InvokerFunc func(ctx context.Context, req *Request, callback CallbackFunc)

func (f InvokerFunc) Invoke(ctx context.Context, req *Request, callback CallbackFunc) {
	f(ctx, req, callback)
}

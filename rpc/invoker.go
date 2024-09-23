package rpc

import (
	"context"

	"google.golang.org/protobuf/proto"
)

type CallbackFunc func(*Response)

func (c CallbackFunc) Response(message proto.Message, err Error) {
	c(&Response{Message: message, Error: err})
}

func (c CallbackFunc) Success(message proto.Message) {
	c(&Response{Message: message})
}

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

type ClientInvoker interface {
	Invoker
	OpenStream(ctx context.Context, req *Request) (*BidiStream, error)
}

type InvokerFunc func(ctx context.Context, req *Request, callback CallbackFunc)

func (f InvokerFunc) Invoke(ctx context.Context, req *Request, callback CallbackFunc) {
	f(ctx, req, callback)
}

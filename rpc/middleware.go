package rpc

import (
	"context"
)

type Interceptor interface {
	Invoke(ctx context.Context, req *Request, callback CallbackFunc, next Invoker)
}

type Middleware struct {
	interceptors []Interceptor
}

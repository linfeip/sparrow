package rpc

import (
	"context"
	"sync"
)

type CallbackFunc func(*Response)

func (c CallbackFunc) Error(err Error) {
	c(&Response{Error: err})
}

type Invoker interface {
	Invoke(ctx context.Context, req *Request, callback CallbackFunc)
}

type InvokerFunc func(ctx context.Context, req *Request, callback CallbackFunc)

func (f InvokerFunc) Invoke(ctx context.Context, req *Request, callback CallbackFunc) {
	f(ctx, req, callback)
}

func NewInvokerChain(finalInvoker Invoker) *InvokerChain {
	return &InvokerChain{finalInvoker: finalInvoker}
}

type InvokerChain struct {
	mu           sync.RWMutex
	finalInvoker Invoker
	interceptors []Interceptor
}

func (c *InvokerChain) AddLast(interceptor Interceptor) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.interceptors = append(c.interceptors, interceptor)
}

func (c *InvokerChain) Invoke(ctx context.Context, req *Request, callback CallbackFunc) {
	chain := c.finalInvoker
	// 包装整个中间件
	for i := len(c.interceptors) - 1; i >= 0; i-- {
		interceptor := c.interceptors[i]
		next := chain
		chain = InvokerFunc(func(ctx context.Context, req *Request, callback CallbackFunc) {
			interceptor.Invoke(ctx, req, callback, next)
		})
	}
	chain.Invoke(ctx, req, callback)
}

package rpc

import "context"

func NewMiddleware() Middleware {
	return &middleware{}
}

type Middleware interface {
	AddLast(interceptors ...Interceptor)
	Build(invoker Invoker) Invoker
}

type middleware struct {
	interceptors []Interceptor
}

func (m *middleware) AddLast(interceptors ...Interceptor) {
	m.interceptors = append(m.interceptors, interceptors...)
}

func (m *middleware) Build(invoker Invoker) Invoker {
	var chain Invoker
	chain = invoker
	// 包装整个中间件
	for i := len(m.interceptors) - 1; i >= 0; i-- {
		interceptor := m.interceptors[i]
		next := chain
		chain = InvokerFunc(func(ctx context.Context, req *Request, callback CallbackFunc) {
			interceptor.Invoke(ctx, req, callback, next)
		})
	}
	return chain
}

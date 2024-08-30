package rpc

import "sparrow/registry"

type Consumer interface {
	Invoker
}

type consumer struct {
	discover registry.Discover
}

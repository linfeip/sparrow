package rpc

import (
	"context"
	"errors"
	"sync"

	"sparrow/network"
	"sparrow/registry"
)

var KeySelectedAddr = "KeySelectedAddr"

func NewClient(opts ...ClientOption) (*Client, error) {
	cliOpts := &ClientOptions{
		scheme: network.TransportTCP,
	}
	for _, opt := range opts {
		opt(cliOpts)
	}

	client := &Client{
		ClientOptions: cliOpts,
		middleware:    NewMiddleware(),
		invokers:      make(map[string]Invoker),
	}

	if len(client.addr) != 0 && client.discover != nil {
		return nil, errors.New("discover or addr only choice one")
	}

	return client, nil
}

type ClientOptions struct {
	discover registry.Discover
	addr     string
	scheme   string
}

type ClientOption func(*ClientOptions)

func WithClientDiscover(discover registry.Discover) ClientOption {
	return func(opts *ClientOptions) {
		opts.discover = discover
	}
}

func WithClientAddr(addr string) ClientOption {
	return func(opts *ClientOptions) {
		opts.addr = addr
	}
}

func WithScheme(scheme string) ClientOption {
	return func(opts *ClientOptions) {
		opts.scheme = scheme
	}
}

type Client struct {
	*ClientOptions
	rw         sync.RWMutex
	middleware Middleware
	invokers   map[string]Invoker
	client     *network.Client
}

func (c *Client) AddLast(interceptors ...Interceptor) {
	c.rw.Lock()
	defer c.rw.Unlock()
	//c.middleware.AddLast(interceptors...)
	//c.invoker = c.middleware.Build(c.invoker).(ClientInvoker)
}

func (c *Client) Invoke(ctx context.Context, req *Request, callback CallbackFunc) {
	invoker, err := c.selectInvoker(ctx, req)
	if err != nil {
		callback(&Response{Error: WrapError(err)})
		return
	}
	invoker.Invoke(ctx, req, callback)
}

func (c *Client) selectInvoker(ctx context.Context, req *Request) (Invoker, error) {
	addr := c.addr
	if len(addr) == 0 {
		// 判断是否引入了注册中心
		node, err := c.discover.Select(ctx, req.Method.ServiceName)
		if err != nil {
			return nil, err
		}
		addr = node.Address
	}

	addr = c.scheme + "://" + addr

	var invoker Invoker
	c.rw.RLock()
	invoker = c.invokers[addr]
	if invoker != nil {
		c.rw.RUnlock()
		return invoker, nil
	}
	c.rw.RUnlock()
	c.rw.Lock()
	defer c.rw.Unlock()

	// double check
	invoker = c.invokers[addr]
	if invoker != nil {
		return invoker, nil
	}

	connection, err := c.client.Connect(addr, network.WithHandler(&Codec{}, &ClientHandler{}))
	if err != nil {
		return nil, err
	}

	inv := &clientInvoker{connection: connection}
	c.invokers[addr] = inv
	return inv, nil
}

func (c *Client) OpenStream(ctx context.Context, req *Request) (*BidiStream, error) {
	panic("implement me")
}

type clientInvoker struct {
	connection *network.Connection
	nextId     uint64
}

func (x *clientInvoker) Invoke(ctx context.Context, req *Request, callback CallbackFunc) {
	x.connection.Write(&PendingRequest{Request: req, Handler: callback})
}

func (x *clientInvoker) OpenStream(ctx context.Context, req *Request) (*BidiStream, error) {
	//TODO implement me
	panic("implement me")
}

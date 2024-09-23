package rpc

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"

	"golang.org/x/net/http2"
	"google.golang.org/protobuf/proto"
	"sparrow/registry"
)

var KeySelectedAddr = "KeySelectedAddr"

func NewClient(opts ...ClientOption) (*Client, error) {
	cliOpts := &ClientOptions{}
	for _, opt := range opts {
		opt(cliOpts)
	}

	client := &Client{
		ClientOptions: cliOpts,
		middleware:    NewMiddleware(),
	}

	if len(client.addr) != 0 && client.discover != nil {
		return nil, errors.New("discover or addr only choice one")
	}

	if client.invoker == nil {
		return nil, errors.New("invoker is required")
	}

	client.invoker = client.middleware.Build(client.invoker).(ClientInvoker)

	return client, nil
}

type ClientOptions struct {
	discover registry.Discover
	addr     string
	invoker  ClientInvoker
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

func WithClientInvoker(invoker ClientInvoker) ClientOption {
	return func(opts *ClientOptions) {
		opts.invoker = invoker
	}
}

type Client struct {
	*ClientOptions
	mu         sync.Mutex
	middleware Middleware
}

func (c *Client) AddLast(interceptors ...Interceptor) {
	c.mu.Lock()
	defer c.mu.Unlock()
	//c.middleware.AddLast(interceptors...)
	//c.invoker = c.middleware.Build(c.invoker).(ClientInvoker)
}

func (c *Client) Invoke(ctx context.Context, req *Request, callback CallbackFunc) {
	addr := c.addr
	if len(addr) == 0 {
		// 判断是否引入了注册中心
		node, err := c.discover.Select(ctx, req.Method.ServiceName)
		if err != nil {
			callback.Error(WrapError(err))
			return
		}
		addr = node.Address
	}
	ctx = context.WithValue(ctx, KeySelectedAddr, addr)
	c.invoker.Invoke(ctx, req, callback)
}

func (c *Client) OpenStream(ctx context.Context, req *Request) (*BidiStream, error) {
	return c.invoker.OpenStream(ctx, req)
}

func NewH2ClientInvoker() ClientInvoker {
	httpClient := &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLSContext: func(_ context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
				return net.Dial(network, addr)
			},
		},
	}
	return &H2ClientInvoker{
		client: httpClient,
	}
}

type H2ClientInvoker struct {
	client *http.Client
}

func (h *H2ClientInvoker) OpenStream(ctx context.Context, req *Request) (*BidiStream, error) {
	var rspReady = make(chan struct{}, 1)
	var rsp *http.Response
	requestReader, requestWriter := io.Pipe()
	stream := NewBidiStream(req.Method.CallType, req.Method.NewOutput)
	stream.SetWriter(requestWriter)
	stream.SetReady(rspReady)
	//addr := ctx.Value(KeySelectedAddr).(string)
	addr := "127.0.0.1:1230"
	url := fmt.Sprintf("http://%s/%s/%s", addr, req.Method.ServiceName, req.Method.MethodName)
	makeRequest := func() {
		defer close(rspReady)
		var httpReq *http.Request
		var err error
		httpReq, err = http.NewRequest("POST", url, requestReader)
		if err != nil {
			panic(err)
		}

		rsp, err = h.client.Do(httpReq)
		if err != nil {
			panic(err)
		}
		stream.SetReader(rsp.Body)
	}
	go makeRequest()
	return stream, nil
}

func (h *H2ClientInvoker) Invoke(ctx context.Context, req *Request, callback CallbackFunc) {
	stream, err := h.OpenStream(ctx, req)
	if err != nil {
		callback.Error(WrapError(err))
		return
	}
	defer stream.Close()
	err = stream.Send(req.Input.(proto.Message))
	if err != nil {
		callback.Error(WrapError(err))
		return
	}
	msg, err := stream.RecvResponse()
	if err != nil {
		callback.Error(WrapError(err))
		return
	}
	callback.Success(msg)
}

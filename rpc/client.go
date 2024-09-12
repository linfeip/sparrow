package rpc

import (
	"bytes"
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

	client.invoker = client.middleware.Build(client.invoker)

	return client, nil
}

type ClientOptions struct {
	discover registry.Discover
	addr     string
	invoker  Invoker
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

func WithClientInvoker(invoker Invoker) ClientOption {
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
	c.middleware.AddLast(interceptors...)
	c.invoker = c.middleware.Build(c.invoker)
}

func (c *Client) Invoke(ctx context.Context, req *Request, callback CallbackFunc) {
	// TODO: client的invoker从外部传入， 可以实现不同的invoker， 比如http2Invoker， tcpInvoker。。。
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

func NewH2ClientInvoker() Invoker {
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

func (h *H2ClientInvoker) Invoke(ctx context.Context, req *Request, callback CallbackFunc) {
	addr := ctx.Value(KeySelectedAddr).(string)
	var data []byte
	var err error
	switch input := req.Input.(type) {
	case proto.Message:
		data, err = proto.Marshal(input)
		if err != nil {
			callback(&Response{
				Error: WrapError(err),
			})
			return
		}
	case []byte:
		data = input
	default:
		callback(&Response{
			Error: WrapError(errors.New("not support input message type")),
		})
		return
	}
	url := fmt.Sprintf("http://%s/%s/%s", addr, req.Method.ServiceName, req.Method.MethodName)
	httpRequest, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	rsp, err := h.client.Do(httpRequest)
	if err != nil {
		callback.Error(WrapError(err))
		return
	}
	defer rsp.Body.Close()
	if rsp.StatusCode == http.StatusOK {
		out, err := io.ReadAll(rsp.Body)
		if err != nil {
			callback(&Response{
				Error: WrapError(err),
			})
		}
		output := req.Method.NewOutput().(proto.Message)
		err = proto.Unmarshal(out, output)
		if err != nil {
			callback.Error(WrapError(err))
			return
		}
		callback.Success(output)
	} else {
		callback.Error(NewError(int32(rsp.StatusCode), errors.New(rsp.Header.Get(ErrHeaderKey))))
	}
}

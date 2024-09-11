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

	"golang.org/x/net/http2"
	"google.golang.org/protobuf/proto"
	"sparrow/registry"
)

func NewClient(opts ...ClientOption) *Client {
	httpClient := &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLSContext: func(_ context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
				return net.Dial(network, addr)
			},
		},
	}

	cliOpts := &ClientOptions{}
	for _, opt := range opts {
		opt(cliOpts)
	}

	client := &Client{
		clientOpts: cliOpts,
		client:     httpClient,
	}

	return client
}

type ClientOptions struct {
	discover registry.Discover
	addr     string
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

type Client struct {
	clientOpts *ClientOptions
	client     *http.Client
}

func (c *Client) Invoke(ctx context.Context, req *Request, callback CallbackFunc) {
	addr := c.clientOpts.addr
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
	rsp, err := c.client.Do(httpRequest)
	if err != nil {
		callback(&Response{
			Error: WrapError(err),
		})
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
			callback(&Response{
				Error: WrapError(err),
			})
			return
		}
		callback(&Response{
			Message: output,
		})
	} else {
		callback(&Response{
			Error: NewError(int32(rsp.StatusCode), errors.New(rsp.Header.Get(ErrHeaderKey))),
		})
	}
}

package network

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
)

func NewClient() *Client {
	return &Client{}
}

type Client struct {
}

func (c *Client) Connect(addr string, options ...Option) (*Connection, error) {
	ur, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	opts := &Options{
		ctx: context.Background(),
	}
	for _, opt := range options {
		if err = opt(opts); err != nil {
			return nil, err
		}
	}

	switch strings.ToLower(ur.Scheme) {
	case "tcp":
		conn, err := net.Dial("tcp", ur.Host)
		if err != nil {
			return nil, err
		}
		connection := NewConnection(opts.ctx, conn, opts.attachment, opts.handlers...)
		go connection.Serve()
		return connection, nil
	default:
		return nil, fmt.Errorf("unsupported scheme: %s", ur.Scheme)
	}
}

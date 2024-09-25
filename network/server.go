package network

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
)

type Server interface {
	Serve(addr string, opts ...Option) error
	ServeAsync(addr string, callback func(err error), opts ...Option)
	Close() error
}

func NewServer() Server {
	return &server{}
}

type server struct {
	opts     *Options
	addr     string
	listener net.Listener
}

func (s *server) ServeAsync(addr string, callback func(err error), opts ...Option) {
	go func() {
		err := s.Serve(addr, opts...)
		callback(err)
	}()
}

func (s *server) Serve(addr string, opts ...Option) error {
	ur, err := url.Parse(addr)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.opts = &Options{
		ctx:    ctx,
		cancel: cancel,
	}
	for _, opt := range opts {
		if err := opt(s.opts); err != nil {
			return err
		}
	}

	switch strings.ToLower(ur.Scheme) {
	case "tcp":
		lis, err := net.Listen("tcp", ur.Host)
		if err != nil {
			return err
		}
		s.listener = lis
		for {
			conn, err := lis.Accept()
			if err != nil {
				return err
			}
			connection := NewConnection(s.opts.ctx, conn, s.opts.attachment, s.opts.handlers...)
			connection.Serve()
		}
	default:
		return fmt.Errorf("unsupported scheme: %s", ur.Scheme)
	}
}

func (s *server) Close() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func AssertLength(n int, err error) int {
	if err != nil {
		panic(err)
	}
	return n
}

func Assert(err error) {
	if err != nil {
		panic(err)
	}
}

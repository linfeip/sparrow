package rpc

import (
	"context"
	"fmt"
	"net"
	"strings"

	"sparrow/network"
	"sparrow/registry"
)

type Server interface {
	ServeAsync() error
	MustRegister(invoker ServiceInvoker)
}

type options struct {
	addr     string
	exporter string
	registry registry.Registry
}

func WithAddress(addr string) Option {
	return func(o *options) {
		o.addr = addr
	}
}

func WithExporter(exporter string) Option {
	return func(o *options) {
		o.exporter = exporter
	}
}

func WithRegistry(r registry.Registry) Option {
	return func(o *options) {
		o.registry = r
	}
}

func (o *options) apply(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

type Option func(*options)

func NewServer(opts ...Option) Server {
	srv := &server{
		opts: &options{},
	}
	srv.opts.apply(opts...)
	srv.sr = NewServiceRegistry(context.Background(), srv.Exporter(), srv.opts.registry)
	srv.server = network.NewServer()
	return srv
}

type server struct {
	sr     ServiceRegistry
	opts   *options
	server network.Server
}

func (s *server) ServeAsync() error {
	if len(s.opts.addr) == 0 {
		return fmt.Errorf("server addr is empty")
	}
	svcHdr := &ServerHandler{ServiceRegistry: s.sr, streams: make(map[uint64]*BidiStream)}
	s.server.ServeAsync(s.opts.addr, func(err error) {
		panic(err)
	}, network.WithHandler(&Codec{}, svcHdr))
	return nil
}

func (s *server) MustRegister(service ServiceInvoker) {
	s.sr.MustRegister(service)
}

func (s *server) Exporter() string {
	exporter := s.opts.exporter
	if len(exporter) == 0 {
		idx := strings.Index(s.opts.addr, ":")
		if idx >= 0 {
			exporter = LocalIP() + s.opts.addr[idx:]
		}
	}
	return exporter
}

func LocalIP() string {
	if addrs, err := net.InterfaceAddrs(); nil == err {
		for _, addr := range addrs {
			if ipAddr, ok := addr.(*net.IPNet); ok && ipAddr.IP.IsPrivate() && !ipAddr.IP.IsLoopback() {
				if ipv4 := ipAddr.IP.To4(); nil != ipv4 {
					return ipv4.String()
				}
			}
		}
	}
	return "127.0.0.1"
}

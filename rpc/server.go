package rpc

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"sparrow/logger"
	"sparrow/registry"
)

type Server interface {
	Serve() error
	RegisterService(serviceInfo *ServiceInfo)
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
		opts:      &options{},
		mux:       http.NewServeMux(),
		providers: make(map[string]*ServiceInfo),
	}
	srv.opts.apply(opts...)
	srv.httpSrv = &http.Server{
		Addr:    srv.opts.addr,
		Handler: h2c.NewHandler(srv.mux, &http2.Server{}),
	}
	return srv
}

type server struct {
	opts      *options
	httpSrv   *http.Server
	mux       *http.ServeMux
	providers map[string]*ServiceInfo
}

func (s *server) RegisterService(serviceInfo *ServiceInfo) {
	for _, methodInfo := range serviceInfo.Methods {
		s.mux.Handle("/"+serviceInfo.ServiceName+"/"+methodInfo.MethodName, methodInfo.Handler)
	}
	s.providers[serviceInfo.ServiceName] = serviceInfo
}

func (s *server) Serve() error {
	if len(s.opts.addr) == 0 {
		return fmt.Errorf("server addr is empty")
	}

	// 开启注册
	if s.opts.registry != nil {
		go func() {
			_ = s.doRegister()
			after := time.After(time.Second * 30)
			for {
				select {
				case <-after:
					_ = s.doRegister()
					after = time.After(time.Second * 30)
				}
			}
		}()
	}

	go func() {
		if err := s.httpSrv.ListenAndServe(); err != nil {
			logger.Errorf("addr: %s serve error: %v", s.opts.addr, err)
		}
	}()

	return nil
}

func (s *server) doRegister() error {
	for service := range s.providers {
		addr := s.opts.exporter
		if len(addr) == 0 {
			idx := strings.Index(s.opts.addr, ":")
			if idx < 0 {
				return errors.New("server addr error")
			}
			addr = LocalIP() + s.opts.addr[idx:]
		}
		err := s.opts.registry.Register(service, addr, &registry.NodeMetadata{
			Address:    addr,
			ID:         addr,
			Weight:     1.0,
			UpdateTime: time.Now().Unix(),
		})
		if err != nil {
			logger.Errorf("register error service: %s address: %s", service, addr)
			return err
		}
	}
	return nil
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

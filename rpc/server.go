package rpc

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/proto"
	"sparrow/logger"
	"sparrow/registry"
)

const ErrHeaderKey = "SparrowError"

type Server interface {
	ServeAsync() error
	ServiceRegistry() *ServiceRegistry
	BuildRoutes()
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
		mux:  http.NewServeMux(),
	}
	srv.opts.apply(opts...)
	srv.serviceRegistry = NewServiceRegistry(context.Background(), srv.Exporter(), srv.opts.registry)
	srv.httpSrv = &http.Server{
		Addr:    srv.opts.addr,
		Handler: h2c.NewHandler(srv.mux, &http2.Server{}),
	}
	return srv
}

type server struct {
	opts            *options
	httpSrv         *http.Server
	mux             *http.ServeMux
	serviceRegistry *ServiceRegistry
}

func (s *server) ServeAsync() error {
	if len(s.opts.addr) == 0 {
		return fmt.Errorf("server addr is empty")
	}
	s.BuildRoutes()
	go func() {
		if err := s.httpSrv.ListenAndServe(); err != nil {
			logger.Errorf("addr: %s serve error: %v", s.opts.addr, err)
		}
	}()
	return nil
}

func (s *server) ServiceRegistry() *ServiceRegistry {
	return s.serviceRegistry
}

func (s *server) BuildRoutes() {
	routes := s.serviceRegistry.BuildRoutes()
	for route, method := range routes {
		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 读取请求数据
			data, err := io.ReadAll(r.Body)
			if err != nil {
				s.rpcHandleError(w, err)
				return
			}

			// 序列化请求数据
			var input = method.NewInput().(proto.Message)
			if err = proto.Unmarshal(data, input); err != nil {
				s.rpcHandleError(w, err)
				return
			}

			// 调用invoker链
			method.Invoker.Invoke(r.Context(), &Request{
				Method: method,
				Input:  input,
			}, func(rsp *Response) {
				// 这里需要区分一下, 这里应该属于业务错误
				if rsp.Error != nil {
					s.rpcHandleError(w, rsp.Error)
					return
				}

				// 写入响应结果
				rBytes, err := proto.Marshal(rsp.Message)
				if err != nil {
					s.rpcHandleError(w, err)
					return
				}

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(rBytes)
			})
		})
		s.mux.Handle(route, httpHandler)
	}
}

func (s *server) rpcHandleError(w http.ResponseWriter, err error) {
	code := http.StatusInternalServerError
	switch e := err.(type) {
	case Error:
		code = int(e.Code())
	}
	// 错误结果通过http header头的方式响应
	w.Header().Set(ErrHeaderKey, err.Error())
	w.WriteHeader(code)
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

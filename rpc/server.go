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
		mux:  http.NewServeMux(),
	}
	srv.opts.apply(opts...)
	srv.sr = NewServiceRegistry(context.Background(), srv.Exporter(), srv.opts.registry)
	srv.httpSrv = &http.Server{
		Addr:    srv.opts.addr,
		Handler: h2c.NewHandler(srv.mux, &http2.Server{}),
	}
	return srv
}

type server struct {
	sr      ServiceRegistry
	opts    *options
	httpSrv *http.Server
	mux     *http.ServeMux
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

func (s *server) MustRegister(service ServiceInvoker) {
	s.sr.MustRegister(service)
	s.BuildRoutes()
}

func (s *server) BuildRoutes() {
	methods := s.sr.Methods()
	for route, method := range methods {
		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			stream := NewBidiStream(method.CallType, method.NewInput)
			stream.SetWriter(&Http2ResponseWriter{Writer: w})
			stream.SetReader(r.Body)
			defer stream.Close()
			var input proto.Message
			switch method.CallType {
			case CallType_Request:
				var err error
				input, err = stream.Recv()
				if err != nil {
					_ = stream.SendResponse(nil, WrapError(err))
					return
				}
			case CallType_BidiStream, CallType_ClientStream:
			case CallType_ServerStream:
				var err error
				input, err = stream.Recv()
				if err != nil {
					_ = stream.SendResponse(nil, WrapError(err))
					return
				}
			}
			method.Invoker.Invoke(r.Context(), &Request{
				Method: method,
				Input:  input,
				Stream: stream,
			}, func(rsp *Response) {
				_ = stream.SendResponse(rsp.Message, rsp.Error)
			})
		})
		s.mux.Handle(route, httpHandler)
	}
}

type Http2ResponseWriter struct {
	io.Writer
}

func (w *Http2ResponseWriter) Write(p []byte) (n int, err error) {
	n, err = w.Writer.Write(p)
	if err != nil {
		return
	}
	if flusher, ok := w.Writer.(http.Flusher); ok {
		flusher.Flush()
	}
	return
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

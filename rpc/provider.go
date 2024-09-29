package rpc

import (
	"context"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
	"sparrow/logger"
	"sparrow/registry"
	"sparrow/utils"
)

type Request struct {
	Method *MethodInfo
	Input  proto.Message
	Stream *BidiStream
}

func WrapService(serviceInfo *ServiceInfo, invoker Invoker) ServiceInvoker {
	return &WrapServiceInvoker{
		serviceInfo: serviceInfo,
		Invoker:     invoker,
	}
}

type WrapServiceInvoker struct {
	serviceInfo *ServiceInfo
	Invoker
}

func (w *WrapServiceInvoker) ServiceInfo() *ServiceInfo {
	return w.serviceInfo
}

type ServiceRegistry interface {
	Register(service ServiceInvoker) error
	MustRegister(service ServiceInvoker)
	ByService(service string) (ServiceInvoker, bool)
	ByRoute(route string) (ServiceInvoker, bool)
}

func NewServiceRegistry(ctx context.Context, exporter string, r registry.Registry) ServiceRegistry {
	sr := &serviceRegistry{
		ctx:        ctx,
		exporter:   exporter,
		registry:   r,
		services:   make(map[string]ServiceInvoker),
		routes:     make(map[string]ServiceInvoker),
		middleware: NewMiddleware(),
	}
	go sr.workLoop()
	return sr
}

type serviceRegistry struct {
	ctx        context.Context
	mu         sync.RWMutex
	services   map[string]ServiceInvoker
	registry   registry.Registry
	exporter   string // 注册的地址
	middleware Middleware
	routes     map[string]ServiceInvoker
}

func (s *serviceRegistry) Register(invoker ServiceInvoker) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	wrapService := s.BuildService(invoker)
	s.services[wrapService.ServiceInfo().ServiceName] = wrapService
	s.makeRoutes(wrapService)
	service := invoker.ServiceInfo()
	err := s.registry.Register(service.ServiceName, s.exporter, &registry.NodeMetadata{
		Address:    s.exporter,
		ID:         s.exporter,
		Weight:     1.0,
		UpdateTime: time.Now().Unix(),
	})
	return err
}

func (s *serviceRegistry) makeRoutes(invoker ServiceInvoker) {
	for _, method := range invoker.ServiceInfo().Methods {
		s.routes[method.Route] = invoker
	}
}

func (s *serviceRegistry) ByService(service string) (ServiceInvoker, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	svc, ok := s.services[service]
	return svc, ok
}

func (s *serviceRegistry) ByRoute(route string) (ServiceInvoker, bool) {
	s.mu.RLock()
	s.mu.RUnlock()
	svc, ok := s.routes[route]
	return svc, ok
}

func (s *serviceRegistry) MustRegister(invoker ServiceInvoker) {
	utils.Assert(s.Register(invoker))
}

func (s *serviceRegistry) BuildService(service ServiceInvoker) ServiceInvoker {
	chain := s.middleware.Build(service)
	return WrapService(service.ServiceInfo(), chain)
}

func (s *serviceRegistry) AddInterceptor(interceptors ...Interceptor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.middleware.AddLast(interceptors...)
}

func (s *serviceRegistry) Methods() map[string]*MethodInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	handlers := make(map[string]*MethodInfo, len(s.services))
	for _, serviceInvoker := range s.services {
		service := serviceInvoker.ServiceInfo()
		for _, method := range service.Methods {
			// build route
			route := "/" + service.ServiceName + "/" + method.MethodName
			method.Invoker = serviceInvoker
			handlers[route] = method
		}
	}
	return handlers
}

func (s *serviceRegistry) workLoop() {
	if s.registry == nil {
		return
	}
	interval := time.Second * 5
	for {
		after := time.After(interval)
		select {
		case <-s.ctx.Done():
			return
		case <-after:
			s.doRegister()
			after = time.After(interval)
		}
	}
}

func (s *serviceRegistry) doRegister() {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, serviceInvoker := range s.services {
		service := serviceInvoker.ServiceInfo()
		err := s.registry.Register(service.ServiceName, s.exporter, &registry.NodeMetadata{
			Address:    s.exporter,
			ID:         s.exporter,
			Weight:     1.0,
			UpdateTime: time.Now().Unix(),
		})
		if err != nil {
			logger.Errorf("register error service: %s address: %s", service.ServiceName, s.exporter)
			return
		}
	}
}

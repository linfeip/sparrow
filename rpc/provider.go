package rpc

import (
	"context"
	"io"
	"sync"
	"time"

	"sparrow/logger"
	"sparrow/registry"
)

type Request struct {
	Method *MethodInfo
	Input  any
	Reader io.ReadCloser
	Writer io.Writer
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

func NewServiceRegistry(ctx context.Context, exporter string, r registry.Registry) *ServiceRegistry {
	serviceRegistry := &ServiceRegistry{
		ctx:        ctx,
		exporter:   exporter,
		registry:   r,
		services:   make(map[string]ServiceInvoker),
		middleware: NewMiddleware(),
	}
	go serviceRegistry.workLoop()
	return serviceRegistry
}

type ServiceRegistry struct {
	ctx        context.Context
	mu         sync.RWMutex
	services   map[string]ServiceInvoker
	registry   registry.Registry
	exporter   string // 注册的地址
	middleware Middleware
}

func (s *ServiceRegistry) Register(service ServiceInvoker) {
	s.mu.Lock()
	defer s.mu.Unlock()
	wrapService := s.BuildService(service)
	s.services[wrapService.ServiceInfo().ServiceName] = wrapService
}

func (s *ServiceRegistry) BuildService(service ServiceInvoker) ServiceInvoker {
	chain := s.middleware.Build(service)
	return WrapService(service.ServiceInfo(), chain)
}

func (s *ServiceRegistry) AddInterceptor(interceptors ...Interceptor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.middleware.AddLast(interceptors...)
}

func (s *ServiceRegistry) BuildRoutes() map[string]*MethodInfo {
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

func (s *ServiceRegistry) workLoop() {
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

func (s *ServiceRegistry) doRegister() {
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

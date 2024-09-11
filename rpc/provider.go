package rpc

import (
	"context"
	"sync"
	"time"

	"sparrow/logger"
	"sparrow/registry"
)

type Request struct {
	Method *MethodInfo
	Input  any
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
		ctx:      ctx,
		exporter: exporter,
		registry: r,
		services: make(map[string]ServiceInvoker),
	}
	go serviceRegistry.workLoop()
	return serviceRegistry
}

type ServiceRegistry struct {
	ctx          context.Context
	mu           sync.RWMutex
	services     map[string]ServiceInvoker
	registry     registry.Registry
	exporter     string // 注册的地址
	interceptors []Interceptor
}

func (s *ServiceRegistry) Register(service ServiceInvoker) {
	s.mu.Lock()
	defer s.mu.Unlock()
	wrapService := s.BuildService(service)
	s.services[wrapService.ServiceInfo().ServiceName] = wrapService
}

func (s *ServiceRegistry) BuildService(service ServiceInvoker) ServiceInvoker {
	// 构建包装中间件后的ServiceInvoker
	var chain Invoker
	chain = service
	// 包装整个中间件
	for i := len(s.interceptors) - 1; i >= 0; i-- {
		interceptor := s.interceptors[i]
		next := chain
		chain = InvokerFunc(func(ctx context.Context, req *Request, callback CallbackFunc) {
			interceptor.Invoke(ctx, req, callback, next)
		})
	}
	return WrapService(service.ServiceInfo(), chain)
}

func (s *ServiceRegistry) AddInterceptor(interceptor Interceptor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.interceptors = append(s.interceptors, interceptor)
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

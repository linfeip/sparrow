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

func NewServiceRegistry(ctx context.Context, exporter string, r registry.Registry) *ServiceRegistry {
	serviceRegistry := &ServiceRegistry{
		ctx:      ctx,
		exporter: exporter,
		registry: r,
		services: make(map[string]*ServiceInfo),
	}
	go serviceRegistry.workLoop()
	return serviceRegistry
}

type ServiceRegistry struct {
	ctx      context.Context
	mu       sync.RWMutex
	services map[string]*ServiceInfo
	registry registry.Registry
	exporter string // 注册的地址
}

func (s *ServiceRegistry) Register(serviceInfo *ServiceInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.services[serviceInfo.ServiceName] = serviceInfo
}

func (s *ServiceRegistry) BuildRoutes() map[string]*MethodInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	handlers := make(map[string]*MethodInfo, len(s.services))
	for _, service := range s.services {
		for _, method := range service.Methods {
			// build route
			route := "/" + service.ServiceName + "/" + method.MethodName
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
	for _, service := range s.services {
		err := s.registry.Register(service.ServiceName, s.exporter, &registry.NodeMetadata{
			Address:    s.exporter,
			ID:         s.exporter,
			Weight:     1.0,
			UpdateTime: time.Now().Unix(),
		})
		if err != nil {
			logger.Errorf("register error service: %s address: %s", service, s.exporter)
			return
		}
	}
}

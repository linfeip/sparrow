package rpc

import "net/http"

type ServiceInfo struct {
	ServiceName string
	Methods     []*MethodInfo
}

type MethodInfo struct {
	MethodName string
	Handler    http.Handler
}

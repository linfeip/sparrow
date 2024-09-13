// Code generated by protoc-gen-sparrow.
// DO NOT EDIT!!!
// source:  echo.proto
package main

import (
	"context"
	"errors"

	"google.golang.org/protobuf/proto"
	"sparrow/rpc"
)

var EchoServiceEchoMethodInfo = &rpc.MethodInfo{
	ServiceName: "main.EchoService",
	MethodName:  "Echo",
	NewInput: func() proto.Message {
		return &EchoRequest{}
	},
	NewOutput: func() proto.Message {
		return &EchoResponse{}
	},
}
var EchoServiceIncrMethodInfo = &rpc.MethodInfo{
	ServiceName: "main.EchoService",
	MethodName:  "Incr",
	NewInput: func() proto.Message {
		return &IncrRequest{}
	},
	NewOutput: func() proto.Message {
		return &IncrResponse{}
	},
}

var EchoServiceServiceInfo = &rpc.ServiceInfo{
	ServiceName: "sample.EchoService",
	Methods: []*rpc.MethodInfo{
		EchoServiceEchoMethodInfo,
		EchoServiceIncrMethodInfo,
	},
}

func NewEchoService(impl EchoService) rpc.ServiceInvoker {
	return &xxxEchoService{
		impl: impl,
	}
}

type xxxEchoService struct {
	impl EchoService
}

func (x *xxxEchoService) Invoke(ctx context.Context, req *rpc.Request, callback rpc.CallbackFunc) {
	switch req.Method.MethodName {
	case "Echo":
		reply, err := x.impl.Echo(ctx, req.Input.(*EchoRequest))
		callback(&rpc.Response{
			Message: reply,
			Error:   rpc.WrapError(err),
		})
	case "Incr":
		reply, err := x.impl.Incr(ctx, req.Input.(*IncrRequest))
		callback(&rpc.Response{
			Message: reply,
			Error:   rpc.WrapError(err),
		})
	default:
		callback(&rpc.Response{
			Error: rpc.NewError(404, errors.New("method not found")),
		})
	}
}

func (x *xxxEchoService) ServiceInfo() *rpc.ServiceInfo {
	return EchoServiceServiceInfo
}

type EchoService interface {
	Echo(ctx context.Context, request *EchoRequest) (*EchoResponse, error)
	Incr(ctx context.Context, request *IncrRequest) (*IncrResponse, error)
}

func NewEchoServiceClient(client *rpc.Client) EchoService {
	return &EchoServiceClient{
		client: client,
	}
}

type EchoServiceClient struct {
	client *rpc.Client
}

func (e *EchoServiceClient) Echo(ctx context.Context, request *EchoRequest) (*EchoResponse, error) {
	rpcRequest := &rpc.Request{
		Method: EchoServiceEchoMethodInfo,
		Input:  request,
	}
	var resp *rpc.Response
	e.client.Invoke(ctx, rpcRequest, func(response *rpc.Response) {
		resp = response
	})
	if resp.Error != nil {
		return nil, resp.Error
	}
	return resp.Message.(*EchoResponse), nil
}
func (e *EchoServiceClient) Incr(ctx context.Context, request *IncrRequest) (*IncrResponse, error) {
	rpcRequest := &rpc.Request{
		Method: EchoServiceIncrMethodInfo,
		Input:  request,
	}
	var resp *rpc.Response
	e.client.Invoke(ctx, rpcRequest, func(response *rpc.Response) {
		resp = response
	})
	if resp.Error != nil {
		return nil, resp.Error
	}
	return resp.Message.(*IncrResponse), nil
}

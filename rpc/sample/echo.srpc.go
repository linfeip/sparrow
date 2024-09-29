// Code generated by protoc-gen-sparrow.
// DO NOT EDIT!!!
// source:  echo.proto
package sample

import (
	"context"
	"errors"

	"google.golang.org/protobuf/proto"
	"sparrow/rpc"
)

var EchoServiceEchoMethodInfo = &rpc.MethodInfo{
	ServiceName: "sample.EchoService",
	MethodName:  "Echo",
	Route:       "/sample.EchoService/Echo",
	CallType:    rpc.CallType_Request,
	NewInput: func() proto.Message {
		return &EchoRequest{}
	},
	NewOutput: func() proto.Message {
		return &EchoResponse{}
	},
}
var EchoServiceIncrMethodInfo = &rpc.MethodInfo{
	ServiceName: "sample.EchoService",
	MethodName:  "Incr",
	Route:       "/sample.EchoService/Incr",
	CallType:    rpc.CallType_Request,
	NewInput: func() proto.Message {
		return &IncrRequest{}
	},
	NewOutput: func() proto.Message {
		return &IncrResponse{}
	},
}

var EchoServicePubsubMethodInfo = &rpc.MethodInfo{
	ServiceName: "sample.EchoService",
	MethodName:  "Pubsub",
	Route:       "/sample.EchoService/Pubsub",
	CallType:    rpc.CallType_BidiStream,
	NewInput: func() proto.Message {
		return &PubsubArgs{}
	},
	NewOutput: func() proto.Message {
		return &PubsubReply{}
	},
}

var EchoServiceClientStreamMethodInfo = &rpc.MethodInfo{
	ServiceName: "sample.EchoService",
	MethodName:  "ClientStream",
	Route:       "/sample.EchoService/ClientStream",
	CallType:    rpc.CallType_BidiStream,
	NewInput: func() proto.Message {
		return &ClientStreamArgs{}
	},
	NewOutput: func() proto.Message {
		return &ClientStreamReply{}
	},
}

var EchoServiceServerStreamMethodInfo = &rpc.MethodInfo{
	ServiceName: "sample.EchoService",
	MethodName:  "ServerStream",
	Route:       "/sample.EchoService/ServerStream",
	CallType:    rpc.CallType_ServerStream,
	NewInput: func() proto.Message {
		return &ServerStreamArgs{}
	},
	NewOutput: func() proto.Message {
		return &ServerStreamReply{}
	},
}

var EchoServiceServiceInfo = &rpc.ServiceInfo{
	ServiceName: "sample.EchoService",
	Methods: map[string]*rpc.MethodInfo{
		EchoServiceEchoMethodInfo.Route:         EchoServiceEchoMethodInfo,
		EchoServiceIncrMethodInfo.Route:         EchoServiceIncrMethodInfo,
		EchoServicePubsubMethodInfo.Route:       EchoServicePubsubMethodInfo,
		EchoServiceClientStreamMethodInfo.Route: EchoServiceClientStreamMethodInfo,
		EchoServiceServerStreamMethodInfo.Route: EchoServiceServerStreamMethodInfo,
	},
}

func NewEchoServiceServer(impl EchoServiceServer) rpc.ServiceInvoker {
	return &xxxEchoService{
		impl: impl,
	}
}

type xxxEchoService struct {
	impl EchoServiceServer
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
	case "Pubsub":
		err := x.impl.Pubsub(ctx, &EchoServicePubsubServer{
			stream: req.Stream,
		})
		callback(&rpc.Response{Error: rpc.WrapError(err)})
	case "ClientStream":
		reply, err := x.impl.ClientStream(ctx, &echoServiceServerClientStream{stream: req.Stream})
		callback.Response(reply, rpc.WrapError(err))
	case "ServerStream":
		err := x.impl.ServerStream(ctx, req.Input.(*ServerStreamArgs), &echoServiceServerServerStream{stream: req.Stream})
		callback.Error(rpc.WrapError(err))
	default:
		callback(&rpc.Response{
			Error: rpc.NewError(404, errors.New("method not found")),
		})
	}
}

func (x *xxxEchoService) ServiceInfo() *rpc.ServiceInfo {
	return EchoServiceServiceInfo
}

type EchoServiceServer interface {
	Echo(ctx context.Context, request *EchoRequest) (*EchoResponse, error)
	Incr(ctx context.Context, request *IncrRequest) (*IncrResponse, error)
	Pubsub(context.Context, IEchoServicePubsubServer) error
	ClientStream(ctx context.Context, stream EchoServiceServerClientStream) (*ClientStreamReply, error)
	ServerStream(ctx context.Context, request *ServerStreamArgs, stream EchoServiceServerServerStream) error
}

type EchoServiceServerServerStream interface {
	Send(msg *ServerStreamReply) error
}

type echoServiceServerServerStream struct {
	stream *rpc.BidiStream
}

func (e *echoServiceServerServerStream) Send(msg *ServerStreamReply) error {
	return e.stream.Send(msg)
}

type EchoServiceServerClientStream interface {
	Recv() (*ClientStreamArgs, error)
}

type echoServiceServerClientStream struct {
	stream *rpc.BidiStream
}

func (e *echoServiceServerClientStream) Recv() (*ClientStreamArgs, error) {
	msg, err := e.stream.Recv()
	if err != nil {
		return nil, err
	}
	return msg.(*ClientStreamArgs), nil
}

type IEchoServicePubsubServer interface {
	Send(*PubsubReply) error
	Recv() (*PubsubArgs, error)
}

type EchoServicePubsubServer struct {
	stream *rpc.BidiStream
}

func (e *EchoServicePubsubServer) Send(reply *PubsubReply) error {
	return e.stream.Send(reply)
}

func (e *EchoServicePubsubServer) Recv() (*PubsubArgs, error) {
	reply, err := e.stream.Recv()
	if err != nil {
		return nil, err
	}
	return reply.(*PubsubArgs), nil
}

type EchoService interface {
	Echo(ctx context.Context, request *EchoRequest) (*EchoResponse, error)
	Incr(ctx context.Context, request *IncrRequest) (*IncrResponse, error)
	Pubsub(ctx context.Context) (IEchoServicePubsubClient, error)
	ClientStream(ctx context.Context) (EchoServiceClientStream, error)
	ServerStream(ctx context.Context, request *ServerStreamArgs) (EchoServiceServerStream, error)
}

type EchoServiceServerStream interface {
	Recv() (*ServerStreamReply, error)
}

type echoServiceServerStream struct {
	stream *rpc.BidiStream
}

func (e *echoServiceServerStream) Recv() (*ServerStreamReply, error) {
	msg, err := e.stream.Recv()
	if err != nil {
		return nil, err
	}
	return msg.(*ServerStreamReply), nil
}

type EchoServiceClientStream interface {
	Send(args *ClientStreamArgs) error
	CloseAndRecv() (*ClientStreamReply, error)
}

type echoServiceClientStream struct {
	stream *rpc.BidiStream
}

func (e *echoServiceClientStream) Send(args *ClientStreamArgs) error {
	return e.stream.Send(args)
}

func (e *echoServiceClientStream) CloseAndRecv() (*ClientStreamReply, error) {
	e.stream.CloseWriter()
	defer e.stream.CloseReader()
	msg, err := e.stream.RecvResponse()
	if err != nil {
		return nil, err
	}
	return msg.(*ClientStreamReply), nil
}

type IEchoServicePubsubClient interface {
	Send(*PubsubArgs) error
	Recv() (*PubsubReply, error)
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
	signal := make(chan struct{})
	e.client.Invoke(ctx, rpcRequest, func(response *rpc.Response) {
		defer close(signal)
		resp = response
	})
	<-signal
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

func (e *EchoServiceClient) Pubsub(ctx context.Context) (IEchoServicePubsubClient, error) {
	req := &rpc.Request{
		Method: EchoServicePubsubMethodInfo,
	}
	stream, err := e.client.OpenStream(ctx, req)
	if err != nil {
		return nil, err
	}
	return &EchoServicePubsubClient{stream: stream}, nil
}

func (e *EchoServiceClient) ClientStream(ctx context.Context) (EchoServiceClientStream, error) {
	req := &rpc.Request{
		Method: EchoServiceClientStreamMethodInfo,
	}
	stream, err := e.client.OpenStream(ctx, req)
	if err != nil {
		return nil, err
	}
	return &echoServiceClientStream{stream: stream}, nil
}

func (e *EchoServiceClient) ServerStream(ctx context.Context, request *ServerStreamArgs) (EchoServiceServerStream, error) {
	req := &rpc.Request{
		Method: EchoServiceServerStreamMethodInfo,
		Input:  request,
	}
	stream, err := e.client.OpenStream(ctx, req)
	if err != nil {
		return nil, err
	}
	err = stream.Send(request)
	if err != nil {
		return nil, err
	}
	return &echoServiceServerStream{stream: stream}, nil
}

type EchoServicePubsubClient struct {
	stream *rpc.BidiStream
}

func (e *EchoServicePubsubClient) Send(args *PubsubArgs) error {
	return e.stream.Send(args)
}

func (e *EchoServicePubsubClient) Recv() (*PubsubReply, error) {
	reply, err := e.stream.Recv()
	if err != nil {
		return nil, err
	}
	return reply.(*PubsubReply), nil
}

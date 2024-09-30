package rpc

import (
	"sync"
	"sync/atomic"

	"google.golang.org/protobuf/proto"
	"sparrow/network"
	"sparrow/utils"
)

type ServerHandler struct {
	ServiceRegistry ServiceRegistry
}

func (s *ServerHandler) HandleRead(ctx network.ReadContext, message any) {
	payload := message.(*ProtoPayload)
	invoker, ok := s.ServiceRegistry.ByRoute(payload.Route)
	if !ok {
		ctx.Connection().Write(ToResponsePayload(payload.StreamId, &Response{Error: ErrNotFound}))
		return
	}

	method := invoker.ServiceInfo().Methods[payload.Route]
	input := method.NewInput()
	err := proto.Unmarshal(payload.GetData(), input)
	if err != nil {
		ctx.Connection().Write(ToResponsePayload(payload.StreamId, &Response{
			Error: WrapError(err),
		}))
		return
	}

	invoker.Invoke(ctx, &Request{
		Method: method,
		Input:  input,
	}, func(response *Response) {
		ctx.Connection().Write(ToResponsePayload(payload.GetStreamId(), response))
	})
}

func ToResponsePayload(id uint64, response *Response) *ProtoPayload {
	rspPayload := &ProtoPayload{
		Type:     CallType_Response,
		StreamId: id,
	}

	if response.Error == nil && response.Message != nil {
		var err error
		rspPayload.Data, err = proto.Marshal(response.Message)
		if err != nil {
			response.Error = WrapError(err)
		}
	}

	if response.Error != nil {
		rspPayload.Error = &ProtoError{
			ErrCode: response.Error.Code(),
			ErrMsg:  response.Error.Error(),
		}
	}

	return rspPayload
}

type ClientHandler struct {
	ctx      network.HandlerContext
	nextId   uint64
	pendings sync.Map
}

func (c *ClientHandler) HandleRead(ctx network.ReadContext, message any) {
	payload := message.(*ProtoPayload)
	v, ok := c.pendings.Load(payload.StreamId)
	if !ok {
		// ignore
		return
	}
	pending := v.(*PendingRequest)
	var out = pending.Request.Method.NewOutput()
	err := proto.Unmarshal(payload.GetData(), out)
	var rsp = &Response{}
	if err != nil {
		rsp.Error = WrapError(err)
		pending.Handler(rsp)
		return
	}
	rsp.Message = out
	_ = utils.GoPool.Submit(func() {
		pending.Handler(rsp)
	})
}

func (c *ClientHandler) HandleWrite(ctx network.WriteContext, message any) {
	pending := message.(*PendingRequest)
	data, err := proto.Marshal(pending.Request.Input)
	if err != nil {
		pending.Handler(&Response{Error: WrapError(err)})
		return
	}
	payload := &ProtoPayload{
		Type:     CallType_Request,
		StreamId: atomic.AddUint64(&c.nextId, 1),
		Route:    pending.Request.Method.Route,
		Data:     data,
	}
	c.pendings.Store(payload.StreamId, pending)
	ctx.HandleWrite(payload)
}

type PendingRequest struct {
	Request *Request
	Handler CallbackFunc
}

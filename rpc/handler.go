package rpc

import (
	_ "net/http/pprof"
	"sync"
	"sync/atomic"

	"google.golang.org/protobuf/proto"
	"sparrow/logger"
	"sparrow/network"
	"sparrow/utils"
)

type ServerHandler struct {
	ServiceRegistry ServiceRegistry
	rw              sync.RWMutex
	streams         map[uint64]*BidiStream
}

func (s *ServerHandler) HandleError(ctx network.HandlerContext, err error) {
	ctx.HandleError(err)
}

func (s *ServerHandler) HandleRead(ctx network.ReadContext, message any) {
	payload := message.(*ProtoPayload)
	invoker, ok := s.ServiceRegistry.ByRoute(payload.Route)
	if !ok {
		ctx.Connection().Write(ToResponsePayload(payload.StreamId, &Response{Error: ErrNotFound}))
		return
	}

	method := invoker.ServiceInfo().Methods[payload.Route]
	stream, exists := s.openStream(ctx.Connection(), payload.StreamId, method)

	if payload.GetType() == CallType_StreamClosed {
		stream.Close()
		return
	}
	stream.Push(payload)
	if exists {
		return
	}

	request := &Request{
		Method: method,
		Stream: stream,
	}

	switch payload.GetType() {
	case CallType_Unary, CallType_ServerStream:
		input, err := stream.Recv(method.NewInput)
		utils.Assert(err)
		request.Input = input
	}

	_ = utils.GoPool.Submit(func() {
		invoker.Invoke(ctx, request, func(response *Response) {
			ctx.Connection().Write(ToResponsePayload(payload.GetStreamId(), response))
			s.closeStream(stream)
		})
	})
}

func (s *ServerHandler) openStream(conn *network.Connection, streamId uint64, method *MethodInfo) (*BidiStream, bool) {
	s.rw.RLock()
	if stream, ok := s.streams[streamId]; ok {
		defer s.rw.RUnlock()
		return stream, true
	}
	s.rw.RUnlock()

	s.rw.Lock()
	defer s.rw.Unlock()
	// double check
	if stream, ok := s.streams[streamId]; ok {
		return stream, true
	}

	stream := NewBidiStream(StreamSideServer, conn, streamId, method)
	s.streams[streamId] = stream
	return stream, false
}

func (s *ServerHandler) closeStream(stream *BidiStream) {
	s.rw.Lock()
	defer s.rw.Unlock()
	stream.Close()
	delete(s.streams, stream.streamId)
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
	conn    *network.Connection
	nextId  uint64
	streams map[uint64]*BidiStream
	rw      sync.RWMutex
}

func (c *ClientHandler) HandleError(ctx network.HandlerContext, err error) {
	logger.Errorf("client error: %v", err)
}

func (c *ClientHandler) HandleRead(ctx network.ReadContext, message any) {
	payload := message.(*ProtoPayload)
	c.rw.RLock()
	var stream = c.streams[payload.GetStreamId()]
	c.rw.RUnlock()
	if stream == nil {
		return
	}
	if payload.GetType() == CallType_StreamClosed {
		stream.Close()
		return
	}

	if payload.GetType() == CallType_Response {
		stream.recvResp(payload)
		return
	}

	_ = utils.GoPool.Submit(func() {
		stream.Push(payload)
	})
	ctx.HandleRead(message)
}

func (c *ClientHandler) HandleWrite(ctx network.WriteContext, message any) {
	ctx.HandleWrite(message)
}

func (c *ClientHandler) OpenStream(req *Request) (*BidiStream, error) {
	streamId := atomic.AddUint64(&c.nextId, 1)
	c.rw.Lock()
	defer c.rw.Unlock()
	stream := NewBidiStream(StreamSideClient, c.conn, streamId, req.Method)
	c.streams[streamId] = stream
	return stream, nil
}

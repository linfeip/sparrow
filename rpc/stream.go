package rpc

import (
	"errors"
	"sync/atomic"

	"google.golang.org/protobuf/proto"
	"sparrow/network"
)

type StreamSide int

const (
	StreamSideClient StreamSide = iota
	StreamSideServer
)

func NewBidiStream(side StreamSide, conn *network.Connection, streamId uint64, method *MethodInfo) *BidiStream {
	return &BidiStream{
		side:     side,
		conn:     conn,
		streamId: streamId,
		msgCh:    make(chan *ProtoPayload, 64),
		method:   method,
		rspCh:    make(chan struct{}),
	}
}

type BidiStream struct {
	conn     *network.Connection
	streamId uint64
	msgCh    chan *ProtoPayload
	method   *MethodInfo
	side     StreamSide
	closed   uint32
	rsp      *Response
	rspCh    chan struct{}
}

func (b *BidiStream) Send(msg proto.Message) error {
	payload, err := EncodeMessage(b.method.CallType, b.method.Route, b.streamId, msg)
	if err != nil {
		return err
	}
	b.conn.Write(payload)
	return nil
}

func (b *BidiStream) Recv(newMsg func() proto.Message) (proto.Message, Error) {
	payload, ok := <-b.msgCh
	if !ok {
		return nil, ErrStreamClosed
	}
	return DecodePayload(payload, newMsg)
}

func (b *BidiStream) Push(payload *ProtoPayload) {
	b.msgCh <- payload
}

func (b *BidiStream) Close() {
	if atomic.CompareAndSwapUint32(&b.closed, 0, 1) {
		close(b.msgCh)
		close(b.rspCh)
	}
}

func (b *BidiStream) CloseAndRecv() (proto.Message, Error) {
	b.SendClose()
	<-b.rspCh
	b.Close()
	return b.rsp.Message, b.rsp.Error
}

func (b *BidiStream) SendClose() {
	payload := &ProtoPayload{
		Type:     CallType_StreamClosed,
		Route:    b.method.Route,
		StreamId: b.streamId,
	}
	b.conn.Write(payload)
}

func (b *BidiStream) recvResp(payload *ProtoPayload) {
	defer func() {
		b.Close()
	}()
	if payload.GetError() != nil {
		b.rsp = &Response{Error: NewError(payload.GetError().GetErrCode(), errors.New(payload.GetError().GetErrMsg()))}
		return
	}

	var msg = b.method.NewOutput()
	if err := proto.Unmarshal(payload.GetData(), msg); err != nil {
		b.rsp = &Response{Error: ErrMsgUnmarshall}
		return
	}
	b.rsp = &Response{Message: msg}
}

func (b *BidiStream) Resp() (proto.Message, Error) {
	<-b.rspCh
	return b.rsp.Message, b.rsp.Error
}

func DecodePayload(payload *ProtoPayload, newMsg func() proto.Message) (proto.Message, Error) {
	if payload.GetError() != nil {
		return nil, NewError(payload.GetError().GetErrCode(), errors.New(payload.GetError().GetErrMsg()))
	}
	var msg = newMsg()
	if err := proto.Unmarshal(payload.GetData(), msg); err != nil {
		return nil, ErrMsgUnmarshall
	}
	return msg, nil
}

func EncodeMessage(call CallType, route string, id uint64, msg proto.Message) (*ProtoPayload, error) {
	data, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}
	payload := &ProtoPayload{
		Type:     call,
		Route:    route,
		Data:     data,
		StreamId: id,
	}
	return payload, nil
}

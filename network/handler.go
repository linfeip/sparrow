package network

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"sparrow/utils"
)

type Handler interface {
}

type ReadHandler interface {
	HandleRead(ctx ReadContext, message any)
}

type WriteHandler interface {
	HandleWrite(ctx WriteContext, message any)
}

type ErrorHandler interface {
	HandleError(ctx HandlerContext, err error)
}

type ConnectedHandler interface {
	HandleConnected(ctx HandlerContext)
}

type CloseHandler interface {
	HandleClose(ctx HandlerContext, err error)
}

type headHandler struct {
	lastTime time.Time
}

func (h *headHandler) HandleWrite(ctx WriteContext, message any) {
	connection := ctx.Connection()
	conn := connection.NetConn()
	switch v := message.(type) {
	case []byte:
		utils.AssertLength(conn.Write(v))
	case [][]byte:
		buffer := net.Buffers(v)
		_, err := buffer.WriteTo(conn)
		utils.Assert(err)
	case io.Reader:
		data, err := io.ReadAll(v)
		utils.Assert(err)
		utils.AssertLength(conn.Write(data))
	default:
		panic(fmt.Errorf("unsupported message type: %T", v))
	}
}

type tailHandler struct {
}

type HandlerContext interface {
	context.Context
	Connection() *Connection
	HandleRead(message any)
	HandleWrite(message any)
	HandleError(err error)
}

type ReadContext interface {
	context.Context
	Connection() *Connection
	HandleRead(message any)
}

type WriteContext interface {
	context.Context
	Connection() *Connection
	HandleWrite(message any)
}

func newHandlerContext(connection *Connection, handler Handler, prev, next *handlerContext) *handlerContext {
	return &handlerContext{connection: connection, Context: connection.ctx, prev: prev, next: next, handler: handler}
}

type handlerContext struct {
	connection *Connection
	context.Context
	next    *handlerContext
	prev    *handlerContext
	handler Handler
}

func (hc *handlerContext) Connection() *Connection {
	return hc.connection
}

func (hc *handlerContext) HandleRead(message any) {
	next := hc
	for {
		next = next.next
		// is last
		if next == nil {
			return
		}
		if handler, ok := next.handler.(ReadHandler); ok {
			handler.HandleRead(next, message)
			return
		}
	}
}

func (hc *handlerContext) HandleWrite(message any) {
	prev := hc
	for {
		prev = prev.prev
		// is head
		if prev == nil {
			return
		}
		if handler, ok := prev.handler.(WriteHandler); ok {
			handler.HandleWrite(prev, message)
			return
		}
	}
}

func (hc *handlerContext) HandleConnected() {
	next := hc
	for {
		next = next.next
		if next == nil {
			return
		}
		if handler, ok := next.handler.(ConnectedHandler); ok {
			handler.HandleConnected(next)
			return
		}
	}
}

func (hc *handlerContext) HandleClose(err error) {
	next := hc
	for {
		next = next.next
		if next == nil {
			return
		}
		if handler, ok := next.handler.(CloseHandler); ok {
			handler.HandleClose(next, err)
			return
		}
	}
}

func (hc *handlerContext) HandleError(err error) {
	next := hc
	for {
		next = next.next
		if next == nil {
			return
		}
		if handler, ok := next.handler.(ErrorHandler); ok {
			handler.HandleError(next, err)
			return
		}
	}
}

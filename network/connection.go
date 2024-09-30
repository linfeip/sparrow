package network

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"

	"sparrow/utils"
)

func NewConnection(ctx context.Context, conn net.Conn, attachment any, handlers ...Handler) *Connection {
	xctx, xcancel := context.WithCancel(ctx)
	c := &Connection{
		ctx:        xctx,
		cancel:     xcancel,
		conn:       conn,
		attachment: attachment,
	}
	c.head = newHandlerContext(c, &headHandler{}, nil, nil)
	c.tail = newHandlerContext(c, &tailHandler{}, nil, nil)
	c.head.next = c.tail
	c.tail.prev = c.head
	for _, h := range handlers {
		c.AddLast(h)
	}
	return c
}

type Connection struct {
	ctx        context.Context
	cancel     context.CancelFunc
	attachment any
	conn       net.Conn
	head       *handlerContext
	tail       *handlerContext
	closed     uint32
}

func (c *Connection) Serve() {
	defer func() {
		if err := recover(); err != nil {
			switch e := err.(type) {
			case error:
				c.Close(e)
			default:
				c.Close(fmt.Errorf("%v", e))
			}
		} else {
			c.Close(nil)
		}
	}()
	c.head.HandleConnected()
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			c.fireHandleRead()
		}
	}
}

func (c *Connection) Close(err error) {
	if atomic.CompareAndSwapUint32(&c.closed, 0, 1) {
		_ = c.conn.Close()
		c.cancel()
		c.head.HandleClose(err)
	}
}

func (c *Connection) Context() context.Context {
	return c.ctx
}

func (c *Connection) fireHandleRead() {
	c.head.HandleRead(c.conn)
}

func (c *Connection) Write(message any) {
	defer func() {
		if err := recover(); err != nil {
			switch e := err.(type) {
			case error:
				c.Close(e)
			default:
				c.Close(fmt.Errorf("%v", e))
			}
		}
	}()
	select {
	case <-c.ctx.Done():
		utils.Assert(c.ctx.Err())
	default:
		c.tail.HandleWrite(message)
	}
}

func (c *Connection) NetConn() net.Conn {
	return c.conn
}

func (c *Connection) AddLast(handler Handler) {
	oldPrev := c.tail.prev
	c.tail.prev = newHandlerContext(c, handler, oldPrev, c.tail)
	oldPrev.next = c.tail.prev
}

func (c *Connection) AddFirst(handler Handler) {
	oldNext := c.head.next
	c.head.next = newHandlerContext(c, handler, c.head, oldNext)
	oldNext.prev = c.head.next
}

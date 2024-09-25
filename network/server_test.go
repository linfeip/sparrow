package network

import (
	"encoding/binary"
	"fmt"
	"io"
	"testing"
	"time"
)

func TestNetwork(t *testing.T) {
	srv := NewServer()
	srv.ServeAsync("tcp://:1231", func(err error) {
		if err != nil {
			panic(err)
		}
	}, WithHandler(&codec{}, &logHandler{prefix: "server"}, &serverHandler{}))

	time.Sleep(time.Second)
	client := NewClient()
	connection, err := client.Connect("tcp://127.0.0.1:1231", WithHandler(&codec{}, &logHandler{prefix: "client"}))
	if err != nil {
		panic(err)
	}

	for {
		err = connection.Write([]byte("HelloWorld"))
		if err != nil {
			panic(err)
		}
		time.Sleep(time.Second)
	}
}

type codec struct {
}

func (c *codec) HandleWrite(ctx HandlerContext, message any) {
	data := message.([]byte)
	var totalBytes = [4]byte{}
	binary.LittleEndian.PutUint32(totalBytes[:], uint32(len(data)))
	ctx.HandleWrite([][]byte{totalBytes[:], data})
}

func (c *codec) HandleRead(ctx HandlerContext, message any) {
	reader := message.(io.Reader)
	var totalBytes = [4]byte{}
	Assert(binary.Read(reader, binary.LittleEndian, &totalBytes))
	total := binary.LittleEndian.Uint32(totalBytes[:])

	buffer := make([]byte, total)
	AssertLength(io.ReadFull(reader, buffer))

	ctx.HandleRead(buffer)
}

type logHandler struct {
	prefix string
}

func (e *logHandler) HandleRead(ctx HandlerContext, message any) {
	data := message.([]byte)
	fmt.Printf("prefix: %s %s\n", e.prefix, string(data))
	ctx.HandleRead(message)
}

type serverHandler struct {
	connection *Connection
	n          int32
}

func (s *serverHandler) HandleRead(ctx HandlerContext, message any) {
	s.n++
	// 回写
	Assert(s.connection.Write(message))
	ctx.HandleRead(message)
}

func (s *serverHandler) HandleConnected(ctx HandlerContext) {
	s.connection = ctx.Connection()
}

func (s *serverHandler) HandleClose(ctx HandlerContext, err error) {
	fmt.Println("handle close err: ", err)
}

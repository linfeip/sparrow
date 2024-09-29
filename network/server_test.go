package network

import (
	"encoding/binary"
	"fmt"
	"io"
	"testing"
	"time"

	"sparrow/utils"
)

var testServer Server
var testClient *Client
var testConnection *Connection
var data = []byte("HelloWorld")

func init() {
	testServer = NewServer()
	testServer.ServeAsync("tcp://:1232", func(err error) {
		if err != nil {
			panic(err)
		}
	}, WithHandler(&codec{}))
	time.Sleep(time.Second)
	testClient = NewClient()
	var err error
	testConnection, err = testClient.Connect("tcp://127.0.0.1:1232", WithHandler(&codec{}))
	if err != nil {
		panic(err)
	}
}

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

func BenchmarkEcho(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			testConnection.Write(data)
		}
	})
}

type codec struct {
	buffer []byte
}

func (c *codec) HandleWrite(ctx WriteContext, message any) {
	data := message.([]byte)
	var totalBytes = [4]byte{}
	binary.LittleEndian.PutUint32(totalBytes[:], uint32(len(data)))
	ctx.HandleWrite([][]byte{totalBytes[:], data})
}

func (c *codec) HandleRead(ctx ReadContext, message any) {
	reader := message.(io.Reader)
	var totalBytes = [4]byte{}
	utils.Assert(binary.Read(reader, binary.LittleEndian, &totalBytes))
	total := binary.LittleEndian.Uint32(totalBytes[:])

	if len(c.buffer) < int(total) {
		c.buffer = make([]byte, int(total))
	}
	buffer := c.buffer[:total]
	utils.AssertLength(io.ReadFull(reader, buffer))

	ctx.HandleRead(buffer)
}

type logHandler struct {
	prefix string
}

func (e *logHandler) HandleRead(ctx ReadContext, message any) {
	data := message.([]byte)
	fmt.Printf("prefix: %s %s\n", e.prefix, string(data))
	ctx.HandleRead(message)
}

type serverHandler struct {
	connection *Connection
	n          int32
}

func (s *serverHandler) HandleRead(ctx ReadContext, message any) {
	s.n++
	// 回写
	utils.Assert(s.connection.Write(message))
	ctx.HandleRead(message)
}

func (s *serverHandler) HandleConnected(ctx HandlerContext) {
	s.connection = ctx.Connection()
}

func (s *serverHandler) HandleClose(ctx HandlerContext, err error) {
	fmt.Println("handle close err: ", err)
}

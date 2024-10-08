package sample

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"sparrow/logger"
	"sparrow/registry"
	"sparrow/rpc"
)

var reg registry.Registry
var client *rpc.Client
var backCtx = context.Background()

func clientMiddleware() rpc.Interceptor {
	return rpc.InterceptorFunc(func(ctx context.Context, req *rpc.Request, callback rpc.CallbackFunc, next rpc.Invoker) {
		selectAddr := ctx.Value(rpc.KeySelectedAddr).(string)
		start := time.Now()
		logger.Debugf("service: %s method: %s client select node addr: %s", req.Method.ServiceName, req.Method.MethodName, selectAddr)
		next.Invoke(ctx, req, callback)
		logger.Debugf("client do service: %s method: %s elapsed: %s", req.Method.ServiceName, req.Method.MethodName, time.Since(start))
	})
}

func init() {
	var err error
	reg = newRegistry()
	startServer(reg, "tcp://:1230", "192.168.218.199:1230")
	//startServer(reg, ":1231")
	client, err = rpc.NewClient(
		rpc.WithClientDiscover(reg),
	)
	if err != nil {
		panic(err)
	}
	client.AddLast(clientMiddleware())
	time.Sleep(time.Second)
}

func TestService(t *testing.T) {
	echoClient := NewEchoServiceClient(client)

	var num = 1
	var wg sync.WaitGroup
	for i := 0; i < num; i++ {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			result, err := echoClient.Echo(backCtx, &EchoRequest{
				Message: fmt.Sprintf("HelloWorld_%d", i),
			})
			if err != nil {
				logger.Errorf("request echo error: %v", err)
				return
			}
			logger.Debugf("client: %d send Echo reply: %s", i, result.Message)
		}()
	}

	wg.Wait()
}

func TestStreamService(t *testing.T) {
	cli := NewEchoServiceClient(client)
	stream, err := cli.Pubsub(backCtx)
	if err != nil {
		panic(err)
	}
	for {
		now := time.Now().String()
		err = stream.Send(&PubsubArgs{
			Data: now,
		})
		if err != nil {
			panic(err)
		}
		reply, err := stream.Recv()
		if err != nil {
			panic(err)
		}
		logger.Debugf("pubsub recv: %s", reply.Data)
		time.Sleep(time.Second)
	}
}

func TestClientStreamService(t *testing.T) {
	cli := NewEchoServiceClient(client)
	stream, err := cli.ClientStream(backCtx)
	if err != nil {
		panic(err)
	}
	for i := 0; i < 2; i++ {
		err = stream.Send(&ClientStreamArgs{
			Value: fmt.Sprint(i),
		})
		if err != nil {
			panic(err)
		}
	}
	reply, err := stream.CloseAndRecv()
	if err != nil {
		panic(err)
	}
	logger.Debugf("recv reply: %s", reply.GetValue())
}

func TestServerStreamService(t *testing.T) {
	cli := NewEchoServiceClient(client)
	stream, err := cli.ServerStream(backCtx, &ServerStreamArgs{
		Value: "5",
	})
	if err != nil {
		panic(err)
	}
	for {
		msg, err := stream.Recv()
		if errors.Is(err, rpc.ErrStreamClosed) {
			break
		}
		if err != nil {
			panic(err)
		}
		logger.Debugf("server recv: %s", msg.GetValue())
	}
}

func BenchmarkService(b *testing.B) {
	echoClient := NewEchoServiceClient(client)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := echoClient.Echo(context.Background(), &EchoRequest{
				Message: "HelloWorld",
			})
			if err != nil {
				panic(err)
			}
		}
	})
}

func TestHttp2Stream(t *testing.T) {
	go func() {
		mux := http.DefaultServeMux
		httpSrv := &http.Server{
			Addr:    ":1223",
			Handler: h2c.NewHandler(mux, &http2.Server{}),
		}

		mux.HandleFunc("/bidi", bidiStreamHandler)

		err := httpSrv.ListenAndServe()
		if err != nil {
			panic(err)
		}
	}()

	time.Sleep(time.Second)

	var rsp *http.Response
	requestReader, requestWriter := io.Pipe()

	var client = &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLSContext: func(_ context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
				return net.Dial(network, addr)
			},
		},
	}
	var rspReady = make(chan struct{}, 1)
	makeRequest := func() {
		defer close(rspReady)
		var req *http.Request
		var err error
		req, err = http.NewRequest("POST", "http://127.0.0.1:1223/bidi", requestReader)
		if err != nil {
			panic(err)
		}

		rsp, err = client.Do(req)
		if err != nil {
			panic(err)
		}
	}

	writeBuff := bytes.NewBuffer(make([]byte, 0, 1024))
	send := func(data []byte) {
		size := len(data)
		writeBuff.Grow(size)
		var sizeBytes [4]byte
		binary.LittleEndian.PutUint32(sizeBytes[:], uint32(size))
		writeBuff.Write(sizeBytes[:])
		writeBuff.Write(data)

		_, err := requestWriter.Write(writeBuff.Bytes())
		if err != nil {
			panic(err)
		}

		writeBuff.Reset()
	}

	buffer := make([]byte, 1024)

	recv := func() {
		<-rspReady
		sizeBytes := [4]byte{}
		n, err := rsp.Body.Read(sizeBytes[:])
		if err != nil {
			panic(err)
		}

		var size = binary.LittleEndian.Uint32(sizeBytes[:])
		n, err = rsp.Body.Read(buffer[:size])
		result := string(buffer[:n])
		logger.Debugf("client recv: %s", result)
	}

	go makeRequest()

	_ = recv

	for {
		send([]byte(time.Now().String()))
		recv()
		time.Sleep(time.Second)
	}
}

func bidiStreamHandler(w http.ResponseWriter, r *http.Request) {
	// 启用 HTTP/2
	if r.ProtoMajor != 2 {
		http.Error(w, "HTTP/2 required", http.StatusBadRequest)
		return
	}

	var buffer = make([]byte, 1024)

	writeBuff := bytes.NewBuffer(make([]byte, 0, 1024))
	count := 0
	for {
		sizeBytes := [4]byte{}
		_, err := r.Body.Read(sizeBytes[:])
		if err != nil {
			panic(err)
		}
		size := binary.LittleEndian.Uint32(sizeBytes[:])
		n, err := r.Body.Read(buffer[:size])
		if err != nil {
			panic(err)
		}
		logger.Debugf("server recv: %s", buffer[:n])

		writeBuff.Grow(int(size))
		binary.LittleEndian.PutUint32(sizeBytes[:], size)
		writeBuff.Write(sizeBytes[:])
		writeBuff.Write(buffer[:n])

		_, err = w.Write(writeBuff.Bytes())
		if err != nil {
			panic(err)
		}
		writeBuff.Reset()
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		count++
	}
}

func newRegistry() registry.Registry {
	r, err := registry.New("zk://192.168.0.43:2181",
		registry.WithTimeout(time.Second*10),
		registry.WithNamespace("sparrow"),
		registry.WithSelector(new(registry.RandomSelector)),
	)

	if err != nil {
		logger.Fatal(err)
	}

	return r
}

func startServer(reg registry.Registry, addr, export string) {
	server := rpc.NewServer(
		rpc.WithAddress(addr),
		rpc.WithRegistry(reg),
		rpc.WithExporter(export),
	)
	if err := server.ServeAsync(); err != nil {
		logger.Fatal(err)
	}
	echoService := &service{}
	//sr.AddInterceptor(middleware.AccessLog())
	server.MustRegister(NewEchoServiceServer(echoService))
}

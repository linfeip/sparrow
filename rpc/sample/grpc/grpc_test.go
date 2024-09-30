package grpc

//go:generate ../tools/protoc.exe --proto_path=. --go_opt=paths=source_relative --go_out=./ --go-grpc_out=. --go-grpc_opt=paths=source_relative echo.proto

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var grpcServer *grpc.Server
var echoClient EchoServiceClient

func init() {
	grpcServer = grpc.NewServer()
	RegisterEchoServiceServer(grpcServer, &testGrpcEchoService{})
	lis, err := net.Listen("tcp", ":1222")
	if err != nil {
		panic(err)
	}
	go func() {
		err = grpcServer.Serve(lis)
		if err != nil {
			panic(err)
		}
	}()

	conn, err := grpc.NewClient("127.0.0.1:1222", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	echoClient = NewEchoServiceClient(conn)
}

type testGrpcEchoService struct {
	UnimplementedEchoServiceServer
}

func (t *testGrpcEchoService) Echo(ctx context.Context, request *EchoRequest) (*EchoResponse, error) {
	return &EchoResponse{
		Message: request.GetMessage(),
	}, nil
}

func (t *testGrpcEchoService) Incr(ctx context.Context, request *IncrRequest) (*IncrResponse, error) {
	panic("implement me")
}

func BenchmarkGrpcService(b *testing.B) {

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

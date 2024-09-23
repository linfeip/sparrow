package sample

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type service struct {
}

func (s *service) Incr(ctx context.Context, request *IncrRequest) (*IncrResponse, error) {
	return &IncrResponse{}, nil
}

func (s *service) Echo(ctx context.Context, request *EchoRequest) (*EchoResponse, error) {
	return &EchoResponse{Message: request.GetMessage()}, nil
}

func (s *service) Pubsub(ctx context.Context, server IEchoServicePubsubServer) error {
	for {
		data, err := server.Recv()
		if err != nil {
			return err
		}
		err = server.Send(&PubsubReply{
			Data: data.GetData(),
		})
		if err != nil {
			return err
		}
	}
}

func (s *service) ClientStream(ctx context.Context, stream EchoServiceServerClientStream) (*ClientStreamReply, error) {
	var values []string
	for {
		args, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		values = append(values, args.GetValue())
	}
	return &ClientStreamReply{Value: strings.Join(values, ",")}, nil
}

func (s *service) ServerStream(ctx context.Context, request *ServerStreamArgs, stream EchoServiceServerServerStream) error {
	num, _ := strconv.Atoi(request.GetValue())
	for i := 0; i < num; i++ {
		err := stream.Send(&ServerStreamReply{
			Value: fmt.Sprint(i),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

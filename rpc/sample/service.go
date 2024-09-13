package sample

import (
	"context"
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

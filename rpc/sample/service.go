package sample

import (
	"context"
)

type service struct {
}

func (s *service) Incr(ctx context.Context, request *IncrRequest) (*IncrResponse, error) {
	//logger.Debug("fire Incr message: ", request.GetNum())
	return &IncrResponse{}, nil
}

func (s *service) Echo(ctx context.Context, request *EchoRequest) (*EchoResponse, error) {
	//logger.Debug("fire Echo message: ", request.Message)
	return &EchoResponse{Message: request.GetMessage()}, nil
}

package sample

import (
	"context"

	"sparrow/logger"
)

type service struct {
}

func (s *service) Echo(ctx context.Context, request *EchoRequest) (*EchoResponse, error) {
	logger.Debug("fire Echo message: ", request.Message)
	return &EchoResponse{Message: request.GetMessage()}, nil
}

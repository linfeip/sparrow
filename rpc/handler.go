package rpc

import (
	"google.golang.org/protobuf/proto"
)

type Response struct {
	Message proto.Message
}

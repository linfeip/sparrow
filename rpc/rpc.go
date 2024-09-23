package rpc

import (
	"bytes"
	"sync"

	"github.com/panjf2000/ants/v2"
	"google.golang.org/protobuf/proto"
)

var GoPool, _ = ants.NewPool(1024)

var BufPool = sync.Pool{
	New: func() any {
		return bytes.NewBuffer(make([]byte, 0, 1024))
	},
}

func init() {
	for i := 0; i < 1024; i++ {
		BufPool.Put(BufPool.New())
	}
}

type Error interface {
	error
	Code() int32
}

type serviceErr struct {
	err  error
	code int32
}

func (e *serviceErr) Error() string {
	return e.err.Error()
}

func (e *serviceErr) Code() int32 {
	return e.code
}

func NewError(code int32, err error) Error {
	return &serviceErr{err: err, code: code}
}

func WrapError(err error) Error {
	if err == nil {
		return nil
	}
	switch err.(type) {
	case Error:
		return err.(Error)
	default:
		return NewError(500, err)
	}
}

type ServiceInfo struct {
	ServiceName string
	Methods     []*MethodInfo
}

type MethodInfo struct {
	NewInput    func() proto.Message
	NewOutput   func() proto.Message
	ServiceName string
	MethodName  string
	CallType    CallType
	Invoker     Invoker
}

type Response struct {
	Error   Error
	Message proto.Message
}

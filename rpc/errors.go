package rpc

import (
	"errors"
	"net/http"
)

var (
	ErrNotFound      = NewError(http.StatusNotFound, errors.New("method not found"))
	ErrStreamClosed  = NewError(600, errors.New("stream msg channel is closed"))
	ErrMsgUnmarshall = NewError(601, errors.New("message unmarshall error"))
)

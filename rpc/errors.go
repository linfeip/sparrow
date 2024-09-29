package rpc

import (
	"errors"
	"net/http"
)

var (
	ErrNotFound = NewError(http.StatusNotFound, errors.New("method not found"))
)

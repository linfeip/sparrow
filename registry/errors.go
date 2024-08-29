package registry

import "errors"

var (
	ErrKeyNotFound      = errors.New("key not found")
	ErrParseKey         = errors.New("parse node key error")
	ErrSelectNodesEmpty = errors.New("select nodes is empty")
)

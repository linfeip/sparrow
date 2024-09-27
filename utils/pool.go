package utils

import (
	"bytes"
	"sync"

	"github.com/panjf2000/ants/v2"
)

var GoPool, _ = ants.NewPool(1024)

var ByteBufferPool = sync.Pool{
	New: func() any {
		return bytes.NewBuffer(make([]byte, 0, 1024))
	},
}

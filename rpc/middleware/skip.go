package middleware

import (
	"context"
	"errors"
	"math/rand/v2"

	"sparrow/logger"
	"sparrow/rpc"
)

type Skip struct {
}

func (s *Skip) Invoke(ctx context.Context, req *rpc.Request, callback rpc.CallbackFunc, next rpc.Invoker) {
	n := rand.IntN(10)
	logger.Debugf("do skip n: %d", n)
	if n < 5 {
		logger.Debug("skipped")
		callback.Error(rpc.WrapError(errors.New("skipped")))
		return
	}
	next.Invoke(ctx, req, callback)
}

type DebugInter struct {
	Name string
}

func (d *DebugInter) Invoke(ctx context.Context, req *rpc.Request, callback rpc.CallbackFunc, next rpc.Invoker) {
	logger.Debugf("debug: %s ...", d.Name)
	next.Invoke(ctx, req, callback)
}

package middleware

import (
	"context"
	"time"

	"sparrow/logger"
	"sparrow/rpc"
)

func AccessLog() rpc.Interceptor {
	return rpc.InterceptorFunc(func(ctx context.Context, req *rpc.Request, callback rpc.CallbackFunc, next rpc.Invoker) {
		logger.Debugf("start access log method: %s/%s", req.Method.ServiceName, req.Method.MethodName)
		start := time.Now()
		next.Invoke(ctx, req, func(response *rpc.Response) {
			callback(response)
			logger.Debugf("end access log method: %s/%s elapsed: %s", req.Method.ServiceName, req.Method.MethodName, time.Since(start))
		})
	})
}

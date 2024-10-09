# 简介

极简的微服务框架

# Feature

- [x] 注册中心
- [x] RPC Server
- [ ] 配置中心
- [ ] 熔断器
- [ ] 链路追踪

# Benchmark

## sparrow

```
goos: windows
goarch: amd64
pkg: sparrow/rpc/sample
cpu: Intel(R) Core(TM) i7-7700K CPU @ 4.20GHz
BenchmarkService
BenchmarkService-8   	  402405	     26829 ns/op	    3142 B/op	      41 allocs/op
PASS
```

## gRPC

```
goos: windows
goarch: amd64
pkg: sparrow/rpc/sample/grpc
cpu: Intel(R) Core(TM) i7-7700K CPU @ 4.20GHz
BenchmarkGrpcService
BenchmarkGrpcService-8   	  198654	     62401 ns/op	    9533 B/op	     173 allocs/op
PASS
```
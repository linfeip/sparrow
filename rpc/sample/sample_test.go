package sample

import (
	"context"
	"fmt"
	"testing"
	"time"

	"sparrow/logger"
	"sparrow/registry"
	"sparrow/rpc"
)

func TestService(t *testing.T) {
	reg, err := registry.New("zk://192.168.0.43:2181",
		registry.WithTimeout(time.Second*10),
		registry.WithNamespace("sparrow"),
		registry.WithSelector(new(registry.RandomSelector)),
	)

	if err != nil {
		logger.Fatal(err)
	}

	for i := 0; i < 1; i++ {
		addr := fmt.Sprintf(":123%d", i)
		server := rpc.NewServer(
			rpc.WithAddress(addr),
			rpc.WithRegistry(reg),
			//rpc.WithExporter("192.168.218.199:1234"),
		)
		if err := server.ServeAsync(); err != nil {
			logger.Fatal(err)
		}
		echoService := &service{}
		server.RegisterService(BuildEchoServiceInfo(echoService))
		server.BuildRoutes()
	}

	echoClient := &EchoServiceClient{
		Discover: reg,
	}

	for {
		time.Sleep(3 * time.Second)
		result, err := echoClient.Echo(context.Background(), &EchoRequest{
			Message: "HelloWorld",
		})
		if err != nil {
			logger.Errorf("request echo error: %v", err)
			continue
		}
		logger.Debugf("client send Echo reply: %s", result.Message)
	}
}

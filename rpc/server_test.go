package rpc

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"golang.org/x/net/http2"
	"sparrow/logger"
)

func TestServer(t *testing.T) {
	srv := NewServer(WithAddress(":1234"))
	if err := srv.ServeAsync(); err != nil {
		t.Fatal(err)
	}

	var transport = http2.Transport{
		DialTLSContext: func(_ context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
			return net.Dial(network, addr)
		},
		AllowHTTP: true,
	}

	var client = &http.Client{
		Transport: &transport,
	}

	time.Sleep(time.Second)

	for i := 0; i < 10; i++ {
		go func() {
			fireService("http://127.0.0.1:1234/sample.EchoService", client)
		}()

		go func() {
			fireService("http://127.0.0.1:1234/sample.SampleService", client)
		}()

		time.Sleep(time.Second * 5)
	}
}

func fireService(addr string, client *http.Client) {
	u1, _ := url.Parse(addr)
	httpReq1 := &http.Request{
		URL:    u1,
		Method: http.MethodPost,
	}
	rsp1, err := client.Do(httpReq1)
	if err != nil {
		panic(err)
	}
	defer rsp1.Body.Close()
	data1, _ := io.ReadAll(rsp1.Body)
	logger.Debugf("service response: %s", data1)
}

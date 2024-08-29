package registry

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestRegistry(t *testing.T) {
	r, err := New("zk://192.168.0.43:2181", WithNamespace("flyer"), WithTimeout(time.Second*5))
	if err != nil {
		t.Fatal(err)
	}

	service := "sample.EchoService"
	defer func() {
		err = r.UnregisterAll(service)
		if err != nil {
			Logger.Error("unregister all error: ", err)
			return
		}
		Logger.Info("unregister all success")
	}()

	for i := 0; i < 10; i++ {
		id := fmt.Sprint(i)
		go func() {
			metadata := &NodeMetadata{
				ID:      id,
				Address: "192.168.0.43:2345",
				Weight:  1.0,
				Version: "0.0.0",
			}
			for {
				metadata.UpdateTime = time.Now().Unix()
				err = r.Register(service, id, metadata)
				if err != nil {
					Logger.Errorf("service: %s id: %s register error: %v", service, id, err)
					time.Sleep(time.Second * 10)
					Logger.Debugf("service: %s id: %s begin retry", service, id)
					continue
				}
				Logger.Debugf("register id: %s time: %d", id, metadata.UpdateTime)
				time.Sleep(time.Second * 10)
			}
		}()
	}

	for {
		time.Sleep(time.Second)
		node, err := r.Select(context.Background(), service)
		if err != nil {
			Logger.Error("select node error: ", err)
			continue
		}
		Logger.Infof("select node key: %s id: %s time: %d", node.Key, node.ID, node.NodeMetadata.UpdateTime)
	}
}

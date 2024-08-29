package registry

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"
)

type NodeMetadata struct {
	ID         string  `json:"id"`
	Address    string  `json:"address"`
	Weight     float64 `json:"weight"`
	Version    string  `json:"version"`
	Load       int32   `json:"load"`
	LoadAvg    float32 `json:"loadAvg"`
	UpdateTime int64   `json:"updateTime"`
	Extends    string  `json:"extends,omitempty"`
}

func (n *NodeMetadata) Marshall() ([]byte, error) {
	return json.Marshal(n)
}

func (n *NodeMetadata) Unmarshall(data []byte) error {
	return json.Unmarshal(data, n)
}

type Node struct {
	Key string
	*NodeMetadata
}

type discover struct {
	opts         *options
	store        Store
	rw           sync.RWMutex
	nodes        map[string]*Node
	serviceNodes map[string][]*Node
}

func (d *discover) Service(service, id string) *Node {
	d.rw.RLock()
	defer d.rw.RUnlock()
	nodes := d.serviceNodes[service]
	for _, node := range nodes {
		if node.ID == id {
			return node
		}
	}
	return nil
}

func (d *discover) Services(service string) []*Node {
	d.rw.RLock()
	defer d.rw.RUnlock()
	return d.serviceNodes[service]
}

func (d *discover) Select(ctx context.Context, service string) (*Node, error) {
	d.rw.RLock()
	defer d.rw.RUnlock()
	selector := d.opts.Selector
	return selector.Select(ctx, service, d.serviceNodes[service])
}

func (d *discover) Start() error {
	if ok, _ := d.store.Exists(d.opts.Namespace); !ok {
		// if not exists create namespace dir
		if err := d.store.Set(d.opts.Namespace, nil); err != nil {
			return err
		}
	}
	// /namespace/service/id
	depth := 3
	go func() {
		for {
			done, err := d.store.WatchEvent(d.opts.Context, d.opts.Namespace, depth, func(ev *WatchEvent) {
				d.rw.Lock()
				defer d.rw.Unlock()

				// 获取service name
				_, service, _, err := d.parseNodeKey(ev.Key)
				if err != nil {
					Logger.Error("parse node key error: ", err)
					return
				}

				d.handleEvent(service, ev)
			})

			if err != nil {
				time.Sleep(time.Second)
				Logger.Error("watch event error: %s retry", err)
				continue
			}

			select {
			case <-d.opts.Context.Done():
				return
			case <-done:
				time.Sleep(time.Second)
				Logger.Debug("watch event done begin retry")
			}
		}
	}()

	return nil
}

func (d *discover) handleEvent(service string, ev *WatchEvent) {
	switch ev.EventType {
	case EventNodeChange:
		metadata := &NodeMetadata{}
		if err := json.Unmarshal(ev.Value, metadata); err != nil {
			Logger.Error("node changed unmarshall metadata error: ", err)
			return
		}
		node, ok := d.nodes[ev.Key]
		if !ok {
			node = &Node{
				Key:          ev.Key,
				NodeMetadata: metadata,
			}
			// 插入节点信息
			d.nodes[ev.Key] = node
			d.serviceNodes[service] = append(d.serviceNodes[service], node)
		} else {
			node.NodeMetadata = metadata
		}
	case EventNodeDel:
		delete(d.nodes, ev.Key)
		delete(d.serviceNodes, service)
	}
}

func (d *discover) parseNodeKey(key string) (namespace, service, id string, err error) {
	// /namespace/service/id
	ss := strings.Split(key, "/")
	if len(ss) != 4 {
		err = ErrParseKey
		return
	}
	// 跳过第一个
	ss = ss[1:]
	return ss[0], ss[1], ss[2], nil
}

package registry

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/go-zookeeper/zk"
)

type EventType int32

const (
	EventNodeChange = 1
	EventNodeDel    = 2
)

type KVPair struct {
	Key   string
	Value []byte
}

type WatchEvent struct {
	EventType
	*KVPair
}

type Store interface {
	Get(key string) (*KVPair, error)
	Set(key string, value []byte) error
	Del(key string) error
	Exists(key string) (bool, error)
	WatchEvent(ctx context.Context, path string, depth int, watcher func(ev *WatchEvent)) (chan struct{}, error)
}

func NewZKStore(servers []string, timeout time.Duration) (Store, error) {
	cli, _, err := zk.Connect(servers, timeout)
	if err != nil {
		return nil, err
	}
	return &zkStore{cli: cli}, nil
}

type zkStore struct {
	cli *zk.Conn
}

func (z *zkStore) Get(key string) (*KVPair, error) {
	data, _, err := z.cli.Get(key)
	return &KVPair{Key: key, Value: data}, err
}

func (z *zkStore) Set(key string, value []byte) error {
	key = z.normalize(key)
	_, err := z.cli.Set(key, value, -1)
	if err != nil && errors.Is(err, zk.ErrNoNode) {
		err = z.createFullPath(key)
		if err != nil {
			return err
		}
		_, err = z.cli.Set(key, value, -1)
		return err
	}
	return err
}

func (z *zkStore) Del(key string) error {
	return z.cli.Delete(key, -1)
}

func (z *zkStore) Exists(key string) (bool, error) {
	ok, _, err := z.cli.Exists(key)
	return ok, err
}

func (z *zkStore) WatchEvent(ctx context.Context, path string, depth int, watcher func(ev *WatchEvent)) (done chan struct{}, err error) {
	path = z.normalize(path)
	var pairsCh <-chan []*KVPair
	var pairCh <-chan *KVPair
	if depth > 1 {
		pairsCh, err = z.WatchList(ctx, path)
		if err != nil {
			return
		}
	} else {
		pairCh, err = z.WatchNode(ctx, path)
		if err != nil {
			return
		}
	}

	done = make(chan struct{}, 1)
	go func() {
		defer close(done)
		var lastList []*KVPair
		for {
			select {
			case <-ctx.Done():
				return
			case pairs, ok := <-pairsCh:
				if !ok {
					return
				}
				// 新增的节点
				var addList = z.newPairs(lastList, pairs)
				for _, pair := range addList {
					_, err = z.WatchEvent(ctx, pair.Key, depth-1, watcher)
					if err != nil {
						return
					}
				}
				lastList = pairs
			case pair, ok := <-pairCh:
				// 节点数据变化
				if !ok {
					watcher(&WatchEvent{EventType: EventNodeDel, KVPair: &KVPair{Key: path}})
					return
				}
				watcher(&WatchEvent{EventType: EventNodeChange, KVPair: pair})
			}
		}
	}()

	return
}

func (z *zkStore) WatchList(ctx context.Context, path string) (<-chan []*KVPair, error) {
	path = z.normalize(path)
	pairs, err := z.List(ctx, path)
	if err != nil {
		return nil, err
	}

	watcher := make(chan []*KVPair, 1)
	go func() {
		defer close(watcher)
		select {
		case watcher <- pairs:
		case <-ctx.Done():
			return
		}

		for {
			_, _, evCh, err := z.cli.ChildrenW(path)
			if err != nil {
				if errors.Is(err, zk.ErrNoNode) {
					return
				}
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Second):
					continue
				}
			}

			select {
			case <-ctx.Done():
				return
			case ev := <-evCh:
				switch ev.Type {
				case zk.EventNodeChildrenChanged:
					pairs, err = z.List(ctx, path)
					if err == nil {
						watcher <- pairs
					}
				case zk.EventNotWatching:
					return
				}
			}
		}
	}()

	return watcher, nil
}

func (z *zkStore) WatchNode(ctx context.Context, path string) (<-chan *KVPair, error) {
	path = z.normalize(path)
	pair, err := z.Get(path)
	if err != nil {
		return nil, err
	}

	watcher := make(chan *KVPair, 1)

	go func() {
		defer close(watcher)

		select {
		case <-ctx.Done():
			return
		case watcher <- pair:
		}

		for {
			_, _, evCh, err := z.cli.GetW(path)
			if err != nil {
				if errors.Is(err, zk.ErrNoNode) {
					return
				}

				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Second):
					continue
				}
			}

			select {
			case ev := <-evCh:
				switch ev.Type {
				case zk.EventNodeDataChanged:
					pair, err := z.Get(path)
					if err == nil {
						watcher <- pair
					}
				case zk.EventNotWatching:
					return
				}
			}
		}
	}()

	return watcher, nil
}

func (z *zkStore) List(ctx context.Context, path string) ([]*KVPair, error) {
	path = z.normalize(path)
	keys, _, err := z.cli.Children(path)
	if err != nil {
		if errors.Is(err, zk.ErrNoNode) {
			return nil, ErrKeyNotFound
		}
		return nil, err
	}

	pairs := make([]*KVPair, 0, len(keys))
	for _, key := range keys {
		pair, err := z.Get(path + z.normalize(key))
		if err != nil {
			if errors.Is(err, zk.ErrNoNode) {
				return z.List(ctx, path)
			}
			return nil, err
		}
		pairs = append(pairs, pair)
	}

	return pairs, nil
}

func (z *zkStore) createFullPath(path string) error {
	dirs := strings.Split(path, "/")
	before := ""
	for _, dir := range dirs {
		if len(dir) == 0 {
			continue
		}
		dir = before + "/" + dir
		before = dir
		_, err := z.cli.Create(dir, []byte{}, 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			if !errors.Is(err, zk.ErrNodeExists) {
				return err
			}
		}
	}
	return nil
}

func (z *zkStore) normalize(path string) string {
	return "/" + strings.Trim(path, "/")
}

func (z *zkStore) newPairs(lastPairs, curPairs []*KVPair) []*KVPair {
	var results []*KVPair
	for _, pair := range curPairs {
		if !z.contains(pair, lastPairs) {
			results = append(results, pair)
		}
	}
	return results
}

func (z *zkStore) contains(pair *KVPair, pairs []*KVPair) bool {
	contains := false
	for _, item := range pairs {
		if item.Key == pair.Key {
			contains = true
			break
		}
	}
	return contains
}

package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"sparrow/logger"
)

const (
	SchemeZK = "zk"
)

var Logger = logger.GetLogger()

type MapMetadata map[string]any

func (m MapMetadata) Marshall() ([]byte, error) {
	return json.Marshal(m)
}

func (m MapMetadata) Unmarshall(data []byte) error {
	return json.Unmarshal(data, &m)
}

type Metadata interface {
	Marshall() ([]byte, error)
	Unmarshall(data []byte) error
}

type Register interface {
	Register(service, id string, metadata Metadata) error
	Unregister(service, id string) error
	UnregisterAll(service string) error
}

type Discover interface {
	Service(service, id string) *Node
	Services(service string) []*Node
	Select(ctx context.Context, service string) (*Node, error)
}

type Registry interface {
	Register
	Discover
}

func New(addr string, opts ...Option) (Registry, error) {
	r := &registry{
		opts: &options{},
	}

	if err := r.opts.apply(opts...); err != nil {
		return nil, err
	}

	// 连接zk
	store, err := newStore(addr, r.opts)
	if err != nil {
		return nil, err
	}
	r.store = store

	r.discover = &discover{
		store:        store,
		opts:         r.opts,
		nodes:        make(map[string]*Node),
		serviceNodes: make(map[string][]*Node),
	}
	if !r.opts.DisableDiscover {
		if err = r.discover.Start(); err != nil {
			return nil, err
		}
	}

	return r, nil
}

func newStore(addr string, opts *options) (Store, error) {
	parsed, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	hosts := strings.Split(parsed.Host, ",")
	switch strings.ToLower(parsed.Scheme) {
	case SchemeZK:
		return NewZKStore(hosts, opts.Timeout)
	default:
		return nil, fmt.Errorf("unsupported scheme: %s", parsed.Scheme)
	}
}

type registry struct {
	*discover
	opts  *options
	store Store
}

func (r *registry) Register(service, id string, metadata Metadata) error {
	data, err := metadata.Marshall()
	if err != nil {
		return err
	}
	key := r.buildKey(service, id)
	return r.store.Set(key, data)
}

func (r *registry) Unregister(service, id string) error {
	key := r.buildKey(service, id)
	return r.store.Del(key)
}

func (r *registry) UnregisterAll(service string) error {
	key := fmt.Sprintf("/%s/%s", r.opts.Namespace, service)
	return r.store.Del(key)
}

func (r *registry) buildKey(service, id string) string {
	service = strings.Trim(service, "/")
	id = strings.Trim(id, "/")
	return fmt.Sprintf("/%s/%s/%s", r.opts.Namespace, service, id)
}

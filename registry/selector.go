package registry

import (
	"context"
	"math/rand"
)

type Selector interface {
	Select(ctx context.Context, service string, nodes []*Node) (*Node, error)
}

type SelectorCreator interface {
	Create(service string) Selector
}

type RandomSelector struct {
}

func (r *RandomSelector) Select(_ context.Context, _ string, nodes []*Node) (*Node, error) {
	if len(nodes) == 0 {
		return nil, ErrSelectNodesEmpty
	}
	rd := rand.Intn(len(nodes))
	return nodes[rd], nil
}

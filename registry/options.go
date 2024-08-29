package registry

import (
	"context"
	"time"
)

var (
	defaultTimeout   = time.Second * 10
	defaultNamespace = "sparrow"
	defaultSelector  = new(RandomSelector)
)

type Option func(opts *options) error

type options struct {
	Namespace       string
	Timeout         time.Duration
	DisableDiscover bool
	Context         context.Context
	Selector        Selector
}

func (opts *options) apply(options ...Option) error {
	opts.Timeout = defaultTimeout
	opts.Namespace = defaultNamespace
	opts.Context = context.Background()
	opts.Selector = defaultSelector
	for _, option := range options {
		if err := option(opts); err != nil {
			return err
		}
	}
	return nil
}

func WithNamespace(namespace string) Option {
	return func(opts *options) error {
		opts.Namespace = namespace
		return nil
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(opts *options) error {
		opts.Timeout = timeout
		return nil
	}
}

func WithDisableDiscover() Option {
	return func(opts *options) error {
		opts.DisableDiscover = true
		return nil
	}
}

func WithContext(ctx context.Context) Option {
	return func(opts *options) error {
		opts.Context = ctx
		return nil
	}
}

func WithSelector(selector Selector) Option {
	return func(opts *options) error {
		opts.Selector = selector
		return nil
	}
}

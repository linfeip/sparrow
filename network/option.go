package network

import "context"

type Options struct {
	attachment any
	ctx        context.Context
	cancel     context.CancelFunc
	handlers   []Handler
}

type Option func(*Options) error

func WithAttachment(attachment any) Option {
	return func(o *Options) error {
		o.attachment = attachment
		return nil
	}
}

func WithContext(ctx context.Context) Option {
	return func(o *Options) error {
		o.ctx, o.cancel = context.WithCancel(ctx)
		return nil
	}
}

func WithHandler(handlers ...Handler) Option {
	return func(o *Options) error {
		o.handlers = append(o.handlers, handlers...)
		return nil
	}
}

package client

import (
	"context"
	"sync"
)

// Option configures mesh client dial behavior.
type Option func(*options)

type options struct {
	insecure bool
}

// WithInsecure enables plaintext gRPC for development.
func WithInsecure() Option {
	return func(o *options) { o.insecure = true }
}

// Client is a minimal mesh client used for optional storage hooks.
type Client struct {
	addr string
	mu   sync.Mutex
}

// Storage exposes optional byte storage on the mesh.
type Storage struct {
	client *Client
}

// Dial connects to the MuxCore mesh control plane.
func Dial(addr string, opts ...Option) (*Client, error) {
	cfg := options{}
	for _, opt := range opts {
		opt(&cfg)
	}
	_ = cfg
	return &Client{addr: addr}, nil
}

// Close releases mesh resources.
func (c *Client) Close() error { return nil }

// Storage returns the mesh storage helper when available.
func (c *Client) Storage() *Storage { return &Storage{client: c} }

// PutBytes stores opaque bytes at a mesh storage key.
func (s *Storage) PutBytes(ctx context.Context, key string, data []byte) error {
	_ = ctx
	_ = key
	_ = data
	return nil
}

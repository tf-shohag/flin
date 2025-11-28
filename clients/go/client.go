package flin

import (
	"errors"
	"fmt"
	"time"

	"github.com/skshohagmiah/flin/internal/net"
)

// Client is the unified Flin client with namespaced APIs
type Client struct {
	// Namespaced APIs
	KV     *KVClient
	Queue  *QueueClient
	Stream *StreamClient
	DB     *DBClient

	// Internal connection management
	pool *net.ConnectionPool
	opts *ClientOptions
}

// ClientOptions for creating a new client
type ClientOptions struct {
	// Server address (unified port for all services)
	Address string

	// Connection pool settings
	MinConnections int
	MaxConnections int

	// Connection timeouts
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// DefaultOptions returns default client options
func DefaultOptions(address string) *ClientOptions {
	return &ClientOptions{
		Address:        address,
		MinConnections: 16,
		MaxConnections: 256,
		DialTimeout:    5 * time.Second,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
	}
}

// NewClient creates a new unified Flin client
func NewClient(opts *ClientOptions) (*Client, error) {
	if opts == nil {
		return nil, errors.New("options cannot be nil")
	}

	if opts.Address == "" {
		return nil, errors.New("address is required")
	}

	// Create shared connection pool
	poolOpts := &net.PoolOptions{
		Address:      opts.Address,
		MinSize:      opts.MinConnections,
		MaxSize:      opts.MaxConnections,
		DialTimeout:  opts.DialTimeout,
		ReadTimeout:  opts.ReadTimeout,
		WriteTimeout: opts.WriteTimeout,
		MaxIdleTime:  5 * time.Minute,
		BufferSize:   65536,
	}

	pool, err := net.NewConnectionPool(poolOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	client := &Client{
		pool: pool,
		opts: opts,
	}

	// Initialize namespaced clients
	client.KV = &KVClient{pool: pool}
	client.Queue = &QueueClient{pool: pool}
	client.Stream = &StreamClient{pool: pool}
	client.DB = &DBClient{pool: pool}

	return client, nil
}

// Close closes all connections
func (c *Client) Close() error {
	if c.pool != nil {
		c.pool.Close()
	}
	return nil
}

// Ping checks if the server is reachable
func (c *Client) Ping() error {
	conn, err := c.pool.Get()
	if err != nil {
		return err
	}
	defer c.pool.Put(conn)
	return nil
}

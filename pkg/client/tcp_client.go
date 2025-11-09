package client

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"time"
)

// TCPClient is a simple TCP client for Flin KV server
type TCPClient struct {
	addr   string
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
	mu     sync.Mutex
}

// NewTCP creates a new TCP client
func NewTCP(addr string) (*TCPClient, error) {
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return &TCPClient{
		addr:   addr,
		conn:   conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
	}, nil
}

// reconnect attempts to reconnect to the server
func (c *TCPClient) reconnect() error {
	if c.conn != nil {
		c.conn.Close()
	}

	conn, err := net.DialTimeout("tcp", c.addr, 5*time.Second)
	if err != nil {
		return err
	}

	c.conn = conn
	c.reader = bufio.NewReader(conn)
	c.writer = bufio.NewWriter(conn)
	return nil
}

// Set sets a key-value pair
func (c *TCPClient) Set(key string, value []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Try operation, reconnect once on failure
	for attempt := 0; attempt < 2; attempt++ {
		// Send: SET key value\r\n
		escapedValue := string(value)
		cmd := fmt.Sprintf("SET %s %s\r\n", key, escapedValue)

		if _, err := c.writer.WriteString(cmd); err != nil {
			if attempt == 0 {
				c.reconnect()
				continue
			}
			return fmt.Errorf("write error: %w", err)
		}
		if err := c.writer.Flush(); err != nil {
			if attempt == 0 {
				c.reconnect()
				continue
			}
			return fmt.Errorf("flush error: %w", err)
		}

		// Read response: +OK\r\n or -ERR ...\r\n
		response, err := c.reader.ReadString('\n')
		if err != nil {
			if attempt == 0 {
				c.reconnect()
				continue
			}
			return fmt.Errorf("read error: %w", err)
		}

		if len(response) > 0 && response[0] == '-' {
			return fmt.Errorf("server error: %s", response[1:])
		}

		return nil
	}

	return fmt.Errorf("failed after retries")
}

// Get gets a value by key
func (c *TCPClient) Get(key string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Send: GET key\r\n
	cmd := fmt.Sprintf("GET %s\r\n", key)
	if _, err := c.writer.WriteString(cmd); err != nil {
		return nil, err
	}
	if err := c.writer.Flush(); err != nil {
		return nil, err
	}

	// Read response: $len\r\ndata\r\n or -ERR ...\r\n
	response, err := c.reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	if len(response) > 0 && response[0] == '-' {
		return nil, fmt.Errorf("key not found")
	}

	// Parse bulk string: $5\r\nhello\r\n
	var length int
	if _, err := fmt.Sscanf(response, "$%d\r\n", &length); err != nil {
		return nil, fmt.Errorf("invalid response format")
	}

	// Read the actual data
	data := make([]byte, length)
	if _, err := c.reader.Read(data); err != nil {
		return nil, err
	}

	// Read trailing \r\n
	c.reader.ReadString('\n')

	return data, nil
}

// Delete deletes a key
func (c *TCPClient) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Send: DEL key\r\n
	cmd := fmt.Sprintf("DEL %s\r\n", key)
	if _, err := c.writer.WriteString(cmd); err != nil {
		return err
	}
	if err := c.writer.Flush(); err != nil {
		return err
	}

	// Read response
	response, err := c.reader.ReadString('\n')
	if err != nil {
		return err
	}

	if len(response) > 0 && response[0] == '-' {
		return fmt.Errorf("server error: %s", response[1:])
	}

	return nil
}

// Exists checks if a key exists
func (c *TCPClient) Exists(key string) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Send: EXISTS key\r\n
	cmd := fmt.Sprintf("EXISTS %s\r\n", key)
	if _, err := c.writer.WriteString(cmd); err != nil {
		return false, err
	}
	if err := c.writer.Flush(); err != nil {
		return false, err
	}

	// Read response: :1\r\n or :0\r\n
	response, err := c.reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	var exists int
	if _, err := fmt.Sscanf(response, ":%d\r\n", &exists); err != nil {
		return false, fmt.Errorf("invalid response format")
	}

	return exists == 1, nil
}

// Close closes the connection
func (c *TCPClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

package flin

import (
	"encoding/json"
	"fmt"

	"github.com/skshohagmiah/flin/internal/net"
	"github.com/skshohagmiah/flin/internal/protocol"
)

// StreamClient handles Stream Processing operations
type StreamClient struct {
	pool *net.ConnectionPool
}

// StreamMessage represents a message in a stream
type StreamMessage struct {
	ID        uint64
	Timestamp int64
	Key       string
	Value     []byte
	Partition int
	Offset    uint64
}

// CreateTopic creates a new topic with partitions and retention
func (c *StreamClient) CreateTopic(topic string, partitions int, retentionMs int64) error {
	conn, err := c.pool.Get()
	if err != nil {
		return err
	}
	defer c.pool.Put(conn)

	request := protocol.EncodeSCreateTopicRequest(topic, partitions, retentionMs)
	if err := conn.Write(request); err != nil {
		return err
	}

	return readOKResponse(conn)
}

// Publish publishes a message to a topic
func (c *StreamClient) Publish(topic string, partition int, key string, value []byte) error {
	conn, err := c.pool.Get()
	if err != nil {
		return err
	}
	defer c.pool.Put(conn)

	request := protocol.EncodeSPublishRequest(topic, partition, key, value)
	if err := conn.Write(request); err != nil {
		return err
	}

	return readOKResponse(conn)
}

// Subscribe subscribes a consumer group to a topic
func (c *StreamClient) Subscribe(topic, group, consumer string) error {
	conn, err := c.pool.Get()
	if err != nil {
		return err
	}
	defer c.pool.Put(conn)

	request := protocol.EncodeSSubscribeRequest(topic, group, consumer)
	if err := conn.Write(request); err != nil {
		return err
	}

	return readOKResponse(conn)
}

// Consume consumes messages from a topic
func (c *StreamClient) Consume(topic, group, consumer string, count int) ([]StreamMessage, error) {
	conn, err := c.pool.Get()
	if err != nil {
		return nil, err
	}
	defer c.pool.Put(conn)

	request := protocol.EncodeSConsumeRequest(topic, group, consumer, count)
	if err := conn.Write(request); err != nil {
		return nil, err
	}

	// Read response
	status, payloadLen, err := conn.ReadHeader()
	if err != nil {
		return nil, err
	}

	if status != protocol.StatusOK {
		return nil, fmt.Errorf("server error: status %d", status)
	}

	payload, err := conn.Read(int(payloadLen))
	if err != nil {
		return nil, err
	}

	// Parse messages
	var messages []StreamMessage
	if err := json.Unmarshal(payload, &messages); err != nil {
		return nil, fmt.Errorf("failed to unmarshal messages: %w", err)
	}

	return messages, nil
}

// Commit commits an offset for a consumer group
func (c *StreamClient) Commit(topic, group string, partition int, offset uint64) error {
	conn, err := c.pool.Get()
	if err != nil {
		return err
	}
	defer c.pool.Put(conn)

	request := protocol.EncodeSCommitRequest(topic, group, partition, offset)
	if err := conn.Write(request); err != nil {
		return err
	}

	return readOKResponse(conn)
}

// Unsubscribe removes a consumer from a group
func (c *StreamClient) Unsubscribe(topic, group, consumer string) error {
	conn, err := c.pool.Get()
	if err != nil {
		return err
	}
	defer c.pool.Put(conn)

	request := protocol.EncodeSUnsubscribeRequest(topic, group, consumer)
	if err := conn.Write(request); err != nil {
		return err
	}

	return readOKResponse(conn)
}

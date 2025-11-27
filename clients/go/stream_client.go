package flin

import (
	"encoding/binary"
	"fmt"

	"github.com/skshohagmiah/flin/internal/net"
	"github.com/skshohagmiah/flin/pkg/protocol"
)

// StreamClient handles stream operations
type StreamClient struct {
	pool *net.ConnectionPool
}

// NewStreamClient creates a new stream client
func NewStreamClient(address string, opts *net.PoolOptions) (*StreamClient, error) {
	// Clone options and set address
	poolOpts := *opts
	poolOpts.Address = address

	pool, err := net.NewConnectionPool(&poolOpts)
	if err != nil {
		return nil, err
	}

	return &StreamClient{pool: pool}, nil
}

// Message represents a stream message
type Message struct {
	Topic     string
	Partition int
	Offset    int64
	Key       string
	Value     []byte
	Timestamp int64
}

// Publish publishes a message to a topic
func (c *StreamClient) Publish(topic, key string, value []byte) error {
	conn, err := c.pool.Get()
	if err != nil {
		return err
	}
	defer c.pool.Put(conn)

	// Partition -1 means auto-select
	req := protocol.EncodeSPublishRequest(topic, -1, key, value)
	if err := conn.Write(req); err != nil {
		return err
	}

	return readOKResponse(conn)
}

// PublishToPartition publishes a message to a specific partition
func (c *StreamClient) PublishToPartition(topic string, partition int, key string, value []byte) error {
	conn, err := c.pool.Get()
	if err != nil {
		return err
	}
	defer c.pool.Put(conn)

	req := protocol.EncodeSPublishRequest(topic, partition, key, value)
	if err := conn.Write(req); err != nil {
		return err
	}

	return readOKResponse(conn)
}

// Consume fetches messages from a topic
func (c *StreamClient) Consume(topic, group, consumer string, count int) ([]*Message, error) {
	conn, err := c.pool.Get()
	if err != nil {
		return nil, err
	}
	defer c.pool.Put(conn)

	req := protocol.EncodeSConsumeRequest(topic, group, consumer, count)
	if err := conn.Write(req); err != nil {
		return nil, err
	}

	// Read multi-value response
	values, err := readMultiValueResponse(conn)
	if err != nil {
		return nil, err
	}

	msgs := make([]*Message, len(values))
	for i, val := range values {
		msg, err := decodeMessage(val)
		if err != nil {
			return nil, err
		}
		msg.Topic = topic
		// Partition is not in the value currently, but we can infer it if needed or add it to encoding
		// For now, let's assume the user knows the topic.
		msgs[i] = msg
	}

	return msgs, nil
}

// Commit commits the consumer offset
func (c *StreamClient) Commit(topic, group string, partition int, offset int64) error {
	conn, err := c.pool.Get()
	if err != nil {
		return err
	}
	defer c.pool.Put(conn)

	req := protocol.EncodeSCommitRequest(topic, group, partition, offset)
	if err := conn.Write(req); err != nil {
		return err
	}

	return readOKResponse(conn)
}

// CreateTopic creates a new topic
func (c *StreamClient) CreateTopic(name string, partitions int, retentionMs int64) error {
	conn, err := c.pool.Get()
	if err != nil {
		return err
	}
	defer c.pool.Put(conn)

	req := protocol.EncodeSCreateTopicRequest(name, partitions, retentionMs)
	if err := conn.Write(req); err != nil {
		return err
	}

	return readOKResponse(conn)
}

// Subscribe joins a consumer group
func (c *StreamClient) Subscribe(topic, group, consumer string) error {
	conn, err := c.pool.Get()
	if err != nil {
		return err
	}
	defer c.pool.Put(conn)

	req := protocol.EncodeSSubscribeRequest(topic, group, consumer)
	if err := conn.Write(req); err != nil {
		return err
	}

	return readOKResponse(conn)
}

// Unsubscribe leaves a consumer group
func (c *StreamClient) Unsubscribe(topic, group, consumer string) error {
	conn, err := c.pool.Get()
	if err != nil {
		return err
	}
	defer c.pool.Put(conn)

	req := protocol.EncodeSUnsubscribeRequest(topic, group, consumer)
	if err := conn.Write(req); err != nil {
		return err
	}

	return readOKResponse(conn)
}

// Helper to decode message from byte slice
// Format: [8:offset][8:timestamp][2:keyLen][key][4:valueLen][value]
func decodeMessage(data []byte) (*Message, error) {
	if len(data) < 18 {
		return nil, fmt.Errorf("invalid message data")
	}

	msg := &Message{}
	pos := 0

	msg.Offset = int64(binary.BigEndian.Uint64(data[pos:]))
	pos += 8
	msg.Timestamp = int64(binary.BigEndian.Uint64(data[pos:]))
	pos += 8

	keyLen := int(binary.BigEndian.Uint16(data[pos:]))
	pos += 2
	if len(data) < pos+keyLen+4 {
		return nil, fmt.Errorf("invalid message data")
	}
	msg.Key = string(data[pos : pos+keyLen])
	pos += keyLen

	valueLen := int(binary.BigEndian.Uint32(data[pos:]))
	pos += 4
	if len(data) < pos+valueLen {
		return nil, fmt.Errorf("invalid message data")
	}
	msg.Value = data[pos : pos+valueLen]

	return msg, nil
}

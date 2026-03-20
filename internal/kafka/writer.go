package kafka

import (
	"context"

	kafkago "github.com/segmentio/kafka-go"
)

func NewWriter(broker string, topic string) *kafkago.Writer {
	return &kafkago.Writer{
		Addr:     kafkago.TCP(broker),
		Topic:    topic,
		Balancer: &kafkago.LeastBytes{},
	}
}

func WriteMessage(ctx context.Context, writer *kafkago.Writer, key []byte, value []byte) error {
	return writer.WriteMessages(ctx, kafkago.Message{
		Key:   key,
		Value: value,
	})
}

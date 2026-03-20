package kafka

import kafkago "github.com/segmentio/kafka-go"

func NewReader(broker string, topic string, groupID string) *kafkago.Reader {
	return kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: []string{broker},
		Topic:   topic,
		GroupID: groupID,
	})
}

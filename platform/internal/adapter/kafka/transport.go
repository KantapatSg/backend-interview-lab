package kafka

import (
	"context"
	"encoding/json"

	"github.com/KantapatSg/backend-interview-lab/platform/internal/app"
	"github.com/segmentio/kafka-go"
)

type Producer struct{ writer *kafka.Writer }

func NewProducer(broker, topic string) *Producer {
	return &Producer{writer: &kafka.Writer{Addr: kafka.TCP(broker), Topic: topic, Balancer: &kafka.Hash{}}}
}
func (p *Producer) Close() error { return p.writer.Close() }
func (p *Producer) Publish(ctx context.Context, event app.Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{Key: []byte(event.JobID), Value: payload})
}

type Consumer struct{ reader *kafka.Reader }

func NewConsumer(broker, topic, group string) *Consumer {
	return &Consumer{reader: kafka.NewReader(kafka.ReaderConfig{Brokers: []string{broker}, Topic: topic, GroupID: group, MinBytes: 1, MaxBytes: 10e6})}
}
func (c *Consumer) Close() error { return c.reader.Close() }
func (c *Consumer) Run(ctx context.Context, handle func(context.Context, app.Event) error) error {
	for {
		message, err := c.reader.FetchMessage(ctx)
		if err != nil {
			return err
		}
		var event app.Event
		if err := json.Unmarshal(message.Value, &event); err != nil {
			continue
		} // poison messages need DLQ in Phase 1B.
		if err := handle(ctx, event); err != nil {
			return err
		}
		// Commit only after the DB receipt/state transaction succeeds.
		if err := c.reader.CommitMessages(ctx, message); err != nil {
			return err
		}
	}
}

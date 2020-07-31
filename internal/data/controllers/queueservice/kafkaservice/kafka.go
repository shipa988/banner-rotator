package kafkaservice

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"

	"github.com/shipa988/banner_rotator/internal/domain/entities"
)

const (
	ErrNilWriter = "writer to kafka is nil. Implement writer if you wanna to produce messages via InitWriter() func"
	ErrNilReader = "reader to kafka is nil. Implement reader if you wanna to consume messages via InitReader() func"
	ErrPush      = "could't push message"
	ErrPull      = "could't pull message"
)

var _ entities.EventQueue = (*KafkaManager)(nil)

type KafkaManager struct {
	writer      *kafka.Writer
	reader      *kafka.Reader
	addr, topic string
}

func NewKafkaManager(addr, topic string) *KafkaManager {
	return &KafkaManager{
		addr:  addr,
		topic: topic,
	}
}
func (k *KafkaManager) InitWriter() {
	k.writer = kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{k.addr},
		Topic:    k.topic,
		Balancer: &kafka.LeastBytes{},
	})
}
func (k *KafkaManager) CloseWriter() {
	k.writer.Close()
}
func (k *KafkaManager) InitReader(consumerGroupID string, minBytesRead, maxBytesRead int) {
	k.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{k.addr},
		GroupID:  consumerGroupID,
		Topic:    k.topic,
		MinBytes: minBytesRead,
		MaxBytes: maxBytesRead,
	})
}
func (k *KafkaManager) CloseReader() {
	k.reader.Close()
}
func (k *KafkaManager) Pull(context context.Context, events chan<- entities.Event) (err error) {
	if k.reader == nil {
		return fmt.Errorf(ErrNilReader)
	}

	e := entities.Event{}
	m := kafka.Message{}
	loop := true
	for loop {
		select {
		case <-context.Done():
			loop = false
			break
		default:
			m, err = k.reader.ReadMessage(context)
			if err != nil {
				break
			}
			if err = json.Unmarshal(m.Value, &e); err != nil {
				break
			}
			events <- e
		}
	}

	return errors.Wrap(err, ErrPull)
}

func (k *KafkaManager) Push(event entities.Event) error {
	if k.writer == nil {
		return fmt.Errorf(ErrNilWriter)
	}

	mess, err := json.Marshal(event)
	if err != nil {
		return errors.Wrap(err, ErrPush)
	}
	if err := k.writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(event.EventType),
			Value: mess,
		},
	); err != nil {
		return errors.Wrap(err, ErrPush)
	}
	return nil
}

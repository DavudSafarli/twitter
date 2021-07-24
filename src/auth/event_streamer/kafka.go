package event_streamer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/davudsafarli/twitter/src/auth/auth_manager"
	"github.com/segmentio/kafka-go"
)

const SEARCH_INGESTION_EVENT_TOPIC = "search-ingestion-events"
const SEARCH_INGESTION_CONSUMER_GROUP_NAME = "search-ingestor-group"

type Kafka struct {
	writer *kafka.Writer
	reader *kafka.Reader
}

func NewKafka(brokers []string) Kafka {
	k := Kafka{}
	k.setupPublisher(context.Background(), brokers)
	k.setupConsumer(context.Background(), brokers)
	return k
}

func (k *Kafka) setupPublisher(ctx context.Context, brokers []string) {
	k.writer = &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Balancer: &kafka.Hash{},
	}
}

func (k *Kafka) setupConsumer(ctx context.Context, brokers []string) {
	k.reader = kafka.NewReader(
		kafka.ReaderConfig{
			Brokers:        brokers,
			GroupID:        SEARCH_INGESTION_CONSUMER_GROUP_NAME,
			Topic:          SEARCH_INGESTION_EVENT_TOPIC,
			MinBytes:       1e3, // 1KB
			MaxBytes:       1e6, // 1MB
			MaxWait:        time.Second,
			ReadBackoffMax: time.Second,
			// continue reading where you left off
			StartOffset: kafka.LastOffset,
		},
	)
}

func (k Kafka) PublishSearchIngestEvent(ctx context.Context, event auth_manager.SearchIngestEvent) error {
	value, err := json.Marshal(event)
	if err != nil {
		panic(err)
	}
	key, _ := json.Marshal(event.UserID)
	err = k.writer.WriteMessages(ctx, kafka.Message{
		Topic: SEARCH_INGESTION_EVENT_TOPIC,
		Key:   key,
		Value: value,
	})

	if err != nil {
		return fmt.Errorf("could not Write Messages: %w", err)
	}
	log.Println("sent the message successfully", event)
	return nil
}

func (k Kafka) ConsumeSearchIngestEvent(ctx context.Context, HandlerFn func(event auth_manager.SearchIngestEvent)) io.Closer {
	// start reading messages and calling HandlerFn
	go func() {
		for {
			// TODO: ReadMessage can be replaced with FetchMessage to have more control over offset commits,
			// TODO: Make the consumer HandlerFn idompotent, and commit after calling HandlerFn.
			log.Println("waiting for a message...")
			m, err := k.reader.ReadMessage(context.Background())
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Println(fmt.Errorf("error while waiting for a message: %w", err))
				break
			}
			log.Printf("message at topic/partition/offset %v/%v/%v: %s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))

			// call HandlerFn
			event := auth_manager.SearchIngestEvent{}
			json.Unmarshal(m.Value, &event)
			HandlerFn(event)
		}
	}()

	return k.reader

}

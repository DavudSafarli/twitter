package kafka_go

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/davudsafarli/twitter/src/auth/auth_manager"
	"github.com/segmentio/kafka-go"
)

type Kafka struct {
	Options KafkaOptions
	Writer  *kafka.Writer
	Reader  *kafka.Reader
}

type KafkaOptions struct {
	Brokers                   []string
	UserEventsTopic           string
	UserEventsConsumerGroupID string
}

func NewKafka(Options KafkaOptions) Kafka {
	k := Kafka{
		Options: Options,
	}
	k.setupPublisher()
	k.setupConsumer()
	return k
}

func DeleteTopic(brokerAddr string, topic string) error {
	conn, err := kafka.Dial("tcp", brokerAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return err
	}
	var controllerConn *kafka.Conn
	controllerConn, err = kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return err
	}
	defer controllerConn.Close()
	err = controllerConn.DeleteTopics(topic)
	if err != nil {
		return err
	}
	return nil
}

func (k *Kafka) setupPublisher() {
	k.Writer = &kafka.Writer{
		Addr:         kafka.TCP(k.Options.Brokers...),
		Topic:        k.Options.UserEventsTopic,
		Balancer:     &kafka.Hash{},
		Logger:       log.Default(),
		MaxAttempts:  10,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		RequiredAcks: 1,
	}
}

func (k *Kafka) setupConsumer() {
	k.Reader = kafka.NewReader(
		kafka.ReaderConfig{
			Brokers:        k.Options.Brokers,
			GroupID:        k.Options.UserEventsConsumerGroupID,
			Topic:          k.Options.UserEventsTopic,
			MinBytes:       1e3, // 1KB
			MaxBytes:       1e6, // 1MB
			MaxWait:        time.Second,
			ReadBackoffMax: time.Second,
			// continue reading where you left off
			StartOffset: kafka.FirstOffset,
			Logger:      log.Default(),
		},
	)
}

func (k Kafka) PublishUserEvent(ctx context.Context, event auth_manager.UserEvent) error {
	value, err := json.Marshal(event)
	if err != nil {
		return err
	}
	key, _ := json.Marshal(event.UserID)
	err = k.Writer.WriteMessages(ctx, kafka.Message{
		// Topic: k.Options.UserEventsTopic,
		Key:   key,
		Value: value,
	})
	if err != nil {
		return fmt.Errorf("could not Write Messages: %w", err)
	}
	log.Println("WROTE MESSAGE TO", k.Options.UserEventsTopic)
	return nil
}

func (k Kafka) ConsumeUserEvents(ctx context.Context, HandlerFn func(event auth_manager.UserEvent)) io.Closer {
	// start reading messages and calling HandlerFn
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		wg.Done()
		for {
			// TODO: ReadMessage can be replaced with FetchMessage to have more control over offset commits,
			// TODO: Make the consumer HandlerFn idompotent, and commit after calling HandlerFn.
			log.Println("waiting for a message for topic:", k.Options.UserEventsTopic)
			m, err := k.Reader.ReadMessage(context.Background())
			if err != nil {
				log.Println(fmt.Errorf("consumer quit. error while waiting for a message: %w", err))
				break
			}
			log.Printf("message at topic/partition/offset %v/%v/%v: %s\n", m.Topic, m.Partition, m.Offset, string(m.Key))

			// call HandlerFn
			event := auth_manager.UserEvent{}
			json.Unmarshal(m.Value, &event)
			HandlerFn(event)
		}
	}()
	wg.Wait()
	return k.Reader

}

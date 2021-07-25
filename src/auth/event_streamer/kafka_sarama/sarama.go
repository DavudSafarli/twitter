package kafka_sarama

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/Shopify/sarama"
	"github.com/davudsafarli/twitter/src/auth/auth_manager"
)

type Options struct {
	Brokers                        []string
	SearchIngestionTopic           string
	SearchIngestionConsumerGroupID string
}

type SaramaClient struct {
	Options Options
	Writer  sarama.SyncProducer
	Reader  sarama.ConsumerGroup
}

// NewSarama creates a new KafkaClient using Sarama Go Library
func NewSarama(options Options) (SaramaClient, error) {
	k := SaramaClient{
		Options: options,
	}
	if err := k.setupPublisher(); err != nil {
		return SaramaClient{}, err
	}
	if err := k.setupConsumer(); err != nil {
		return SaramaClient{}, err
	}
	return k, nil
}

// setupPublisher creates Publisher/Producer/Writer
// TODO: expose necessary options to make it customizable by client
func (k *SaramaClient) setupPublisher() error {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll // Wait for all in-sync replicas to ack the message
	config.Producer.Retry.Max = 10                   // Retry up to 10 times to produce the message
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(k.Options.Brokers, config)
	if err != nil {
		return err
	}
	k.Writer = producer
	return err
}

// setupConsumer creates Consumer.
// TODO: expose necessary options to make it customizable by client
func (k *SaramaClient) setupConsumer() error {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	client, err := sarama.NewConsumerGroup(k.Options.Brokers, k.Options.SearchIngestionConsumerGroupID, config)

	if err != nil {
		return err
	}
	k.Reader = client
	return nil
}

// PublishSearchIngestEvent publishes a SearchIngest event
func (k SaramaClient) PublishSearchIngestEvent(ctx context.Context, event auth_manager.SearchIngestEvent) error {
	value := &JSONEncoder{Value: event}
	p, offset, err := k.Writer.SendMessage(&sarama.ProducerMessage{
		Topic: k.Options.SearchIngestionTopic,
		Key:   sarama.StringEncoder(fmt.Sprint(event.UserID)),
		Value: value,
	})
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}
	log.Printf("Message is writter to topic/partition/offset : %v/%v/%v", k.Options.SearchIngestionTopic, p, offset)
	return nil
}

// ConsumeSearchIngestEvent starts a consumer for SearchIngestion events
func (k SaramaClient) ConsumeSearchIngestEvent(ctx context.Context, HandlerFn func(event auth_manager.SearchIngestEvent)) io.Closer {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		wg.Done()
		consumer := SimpleGroupConsumer{
			handlerFn: func(buff []byte) {
				event := auth_manager.SearchIngestEvent{}
				json.Unmarshal(buff, &event)
				HandlerFn(event)
			},
		}
		for {
			if err := k.Reader.Consume(ctx, []string{k.Options.SearchIngestionTopic}, &consumer); err != nil {
				log.Printf("Error from consumer: %v", err)
				return
			}
		}
	}()
	wg.Wait()
	return k.Reader
}

// SimpleGroupConsumer satisfies sarama.ConsumerGroupHandler interface and used for consuming messages from a topic partition.
// It calls the given handlerFn function and commits the message.
type SimpleGroupConsumer struct {
	// TODO: add error return type
	handlerFn func(buf []byte)
}

func (c SimpleGroupConsumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}
func (c SimpleGroupConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}
func (c SimpleGroupConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		log.Printf("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
		c.handlerFn(message.Value)
		session.MarkMessage(message, "")
	}
	return nil
}

// ------- JSON encoder
// JSONEncoder satisfies sarama.Encoder interface and used to publish events.
// Uses json#Marshall to encode the Value
type JSONEncoder struct {
	Value   interface{}
	encoded []byte
	err     error
}

func (e *JSONEncoder) Encode() ([]byte, error) {
	if e.encoded == nil && e.err == nil {
		e.encoded, e.err = json.Marshal(e.Value)
	}
	return e.encoded, e.err
}
func (e *JSONEncoder) Length() int {
	if e.encoded == nil && e.err == nil {
		e.encoded, e.err = json.Marshal(e.Value)
	}
	return len(e.encoded)
}

//
// -- utility functions
func DeleteTopic(brokers []string, topic string) error {
	config := sarama.NewConfig()

	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 10
	config.Producer.Return.Successes = true
	clusterAdmin, err := sarama.NewClusterAdmin(brokers, config)
	if err != nil {
		return err
	}
	return clusterAdmin.DeleteTopic(topic)
}

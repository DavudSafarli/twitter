package kafka_sarama

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/Shopify/sarama"
	"github.com/davudsafarli/twitter/auth"
)

var decoder JSONEncoderDecoder

type Options struct {
	Brokers                   []string
	UserEventsTopic           string
	UserEventsConsumerGroupID string
}

type SaramaClient struct {
	Options Options
	Writer  sarama.SyncProducer
	Reader  sarama.ConsumerGroup

	handlers struct {
		signupEventHandler func(event auth.ConsumedSignupEvent)
	}
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

	client, err := sarama.NewConsumerGroup(k.Options.Brokers, k.Options.UserEventsConsumerGroupID, config)
	if err != nil {
		return err
	}
	k.Reader = client
	return nil
}

// KafkaMessage is the final struct that is encoded and sent to a kafka topic as a value
type KafkaMessage struct {
	PublishedAt     time.Time
	UserSignupEvent auth.SignupEvent `json:",omitempty"`
}

// Timestamp returns the time that kafka message was sent to the kafka topic
func (msg KafkaMessage) Timestamp() time.Time {
	return msg.PublishedAt
}

// SignupEvent returns the currenly consumed SignupEvent
func (msg KafkaMessage) SignupEvent() auth.SignupEvent {
	return msg.UserSignupEvent
}

// PublishUserSignupEvent publishes a UserEvent
func (k SaramaClient) PublishUserSignupEvent(ctx context.Context, event auth.SignupEvent) error {
	msg := KafkaMessage{
		PublishedAt:     time.Now(),
		UserSignupEvent: event,
	}
	value := &JSONEncoderDecoder{Value: msg}
	p, offset, err := k.Writer.SendMessage(&sarama.ProducerMessage{
		Topic: k.Options.UserEventsTopic,
		Key:   sarama.StringEncoder(fmt.Sprint(event.ID)),
		Value: value,
	})
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}
	log.Printf("Message is writter to topic/partition/offset : %v/%v/%v", k.Options.UserEventsTopic, p, offset)
	return nil
}

// UnregisterUserSignupEventConsumer unregisters the handler function for consuming "UserSignupEvent"s
func (k *SaramaClient) UnregisterUserSignupEventConsumer(ctx context.Context, handlerFn func(event auth.ConsumedSignupEvent)) {
	k.handlers.signupEventHandler = nil
}

// RegisterUserSignupEventConsumer registers a handler function for consuming "UserSignupEvent"s
func (k *SaramaClient) RegisterUserSignupEventConsumer(ctx context.Context, handlerFn func(event auth.ConsumedSignupEvent)) {
	k.handlers.signupEventHandler = handlerFn
}

// StartConsume starts listening kafka topics and send messages to the registered consumers.
// It can see the registered handlers, and only listen the topics that are have handlers for
func (k *SaramaClient) StartConsume(ctx context.Context) io.Closer {
	go func() {
		consumer := SimpleGroupConsumer{
			handlerFn: func(msg KafkaMessage) {
				if ok := (msg.UserSignupEvent != auth.SignupEvent{}); ok {
					k.handlers.signupEventHandler(msg)
				}
			},
		}
		for {
			if err := k.Reader.Consume(ctx, []string{k.Options.UserEventsTopic}, &consumer); err != nil {
				log.Printf("Error from consumer: %v", err)
				return
			}
		}
	}()
	return k.Reader
}

// SimpleGroupConsumer satisfies sarama.ConsumerGroupHandler interface and used for consuming messages from a topic partition.
// It calls the given handlerFn function and commits the message.
type SimpleGroupConsumer struct {
	// TODO: add error return type
	handlerFn func(msg KafkaMessage)
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
		v, err := decoder.Decode(message.Value)
		if err != nil {
			log.Printf("sarama failed to decode an incoming kafka message: %v", err)
		}
		c.handlerFn(v)
		session.MarkMessage(message, "")
	}
	return nil
}

// ------- JSON encoder
// JSONEncoderDecoder satisfies sarama.Encoder interface and used to publish events.
// Uses json#Marshall to encode the Value
type JSONEncoderDecoder struct {
	Value   KafkaMessage
	encoded []byte
	err     error
}

func (e *JSONEncoderDecoder) Decode(buf []byte) (KafkaMessage, error) {
	v := KafkaMessage{}
	err := json.Unmarshal(buf, &v)
	return v, err
}

func (e *JSONEncoderDecoder) Encode() ([]byte, error) {
	if e.encoded == nil && e.err == nil {
		e.encoded, e.err = json.Marshal(e.Value)
	}
	return e.encoded, e.err
}

func (e *JSONEncoderDecoder) Length() int {
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

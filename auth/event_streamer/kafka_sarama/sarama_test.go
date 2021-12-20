package kafka_sarama_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/davudsafarli/twitter/auth/contracts"
	"github.com/davudsafarli/twitter/auth/event_streamer/kafka_sarama"
	"github.com/davudsafarli/twitter/auth/test_helpers"
	"github.com/stretchr/testify/require"
)

func TestSarama(t *testing.T) {
	// TODO: Move test-topic creating and DeleteTopic to EventProducerConsumerContract.Subject
	topicName := fmt.Sprintf("sarama-test-%016x", rand.Int63())
	consumerName := fmt.Sprint(topicName, "-consumer")
	sarama, err := kafka_sarama.NewSarama(kafka_sarama.Options{
		Brokers:                   test_helpers.BROKERS,
		UserEventsTopic:           topicName,
		UserEventsConsumerGroupID: consumerName,
	})
	require.Nil(t, err)
	contracts.EventProducerConsumerContract{
		Subject: &sarama,
	}.Test(t)
	t.Cleanup(func() {
		require.Nil(t, kafka_sarama.DeleteTopic(test_helpers.BROKERS, topicName))
	})
}

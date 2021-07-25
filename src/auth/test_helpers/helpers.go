package test_helpers

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/davudsafarli/twitter/src/auth/auth_manager"
	"github.com/davudsafarli/twitter/src/auth/event_streamer/kafka_sarama"
	"github.com/stretchr/testify/require"
)

const DB_CONN_STR = "postgres://postgres:PGPWD123@localhost:7001/postgres?sslmode=disable"

var BROKERS = []string{"localhost:9094"}

const SEARCH_INGESTION_TOPIC = "search-ingestion-events"
const SEARCH_INGESTION_CONSUMER_GROUP_ID = "search-ingestor-group"

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

func HopefullyUniqueUser() auth_manager.User {
	return auth_manager.User{
		Email:    fmt.Sprintf("Email-%X", random.Int()),
		Username: fmt.Sprintf("Username-%X", random.Int()),
		Password: fmt.Sprintf("Password-%X", random.Int()),
	}
}

func GetEventProducerConsumer(t *testing.T) auth_manager.EventProducerConsumer {
	topicName := fmt.Sprintf("topic-for-test-%016x", random.Int63())
	consumerID := fmt.Sprint(topicName, "-consumer")
	k, err := kafka_sarama.NewSarama(kafka_sarama.Options{
		Brokers:                        BROKERS,
		SearchIngestionTopic:           topicName,
		SearchIngestionConsumerGroupID: consumerID,
	})
	require.Nil(t, err)
	t.Cleanup(func() {
		require.Nil(t, kafka_sarama.DeleteTopic(BROKERS, topicName))
	})
	return k
}

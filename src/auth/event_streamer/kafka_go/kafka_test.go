package kafka_go_test

import (
	"testing"

	"github.com/davudsafarli/twitter/src/auth/auth_manager/contracts"
	"github.com/davudsafarli/twitter/src/auth/event_streamer/kafka_go"
	"github.com/davudsafarli/twitter/src/auth/test_helpers"
)

func TestKafka(t *testing.T) {
	t.Skip()
	kafka := kafka_go.NewKafka(kafka_go.KafkaOptions{
		Brokers:                        test_helpers.BROKERS,
		SearchIngestionTopic:           test_helpers.SEARCH_INGESTION_TOPIC,
		SearchIngestionConsumerGroupID: test_helpers.SEARCH_INGESTION_CONSUMER_GROUP_ID,
	})

	contracts.EventProducerConsumerContract{
		Subject: kafka,
	}.Test(t)
}

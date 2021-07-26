package kafka_go_test

import (
	"testing"
)

func TestKafka(t *testing.T) {
	t.Skip()
	// kafka := kafka_go.NewKafka(kafka_go.KafkaOptions{
	// 	Brokers:                   test_helpers.BROKERS,
	// 	UserEventsTopic:           test_helpers.SEARCH_INGESTION_TOPIC,
	// 	UserEventsConsumerGroupID: test_helpers.SEARCH_INGESTION_CONSUMER_GROUP_ID,
	// })

	// contracts.EventProducerConsumerContract{
	// 	Subject: kafka,
	// }.Test(t)
}

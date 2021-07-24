package event_streamer_test

import (
	"testing"

	"github.com/davudsafarli/twitter/src/auth/auth_manager/contracts"
	"github.com/davudsafarli/twitter/src/auth/event_streamer"
	"github.com/davudsafarli/twitter/src/auth/test_helpers"
)

func TestKafka(t *testing.T) {
	kafka := event_streamer.NewKafka(test_helpers.BROKERS)

	contracts.EventProducerConsumerContract{
		Subject: kafka,
	}.Test(t)
}

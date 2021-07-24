package contracts

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/adamluzsi/testcase"
	"github.com/davudsafarli/twitter/src/auth/auth_manager"
	"github.com/stretchr/testify/require"
)

// \/----TEST---\/

type EventProducerConsumerContract struct {
	Subject auth_manager.EventProducerConsumer
}

func (c EventProducerConsumerContract) Test(t *testing.T) {
	t.Run(`Published event will eventually be consumed by Consumer`, func(t *testing.T) {
		data := fmt.Sprint(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(999999))
		buf := []byte(data)
		publishedEvent := auth_manager.SearchIngestEvent{
			UserID: 1,
			Data:   buf,
		}
		// publish event
		require.Nil(t, c.Subject.PublishSearchIngestEvent(context.Background(), publishedEvent))

		consumedEvent := auth_manager.SearchIngestEvent{}
		// start consumer
		consumer := c.Subject.ConsumeSearchIngestEvent(context.Background(), func(event auth_manager.SearchIngestEvent) {
			consumedEvent = event
		})
		t.Cleanup(func() {
			require.Nil(t, consumer.Close())
		})

		r := testcase.Retry{Strategy: testcase.Waiter{WaitTimeout: 2 * time.Second, WaitDuration: time.Second / 3}}
		r.Assert(t, func(tb testing.TB) {
			require.Equal(tb, publishedEvent, consumedEvent)
		})
	})
}

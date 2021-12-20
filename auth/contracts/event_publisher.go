package contracts

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/adamluzsi/testcase"
	"github.com/davudsafarli/twitter/auth"
	"github.com/stretchr/testify/require"
)

// \/----TEST---\/

type EventProducerConsumerContract struct {
	Subject interface {
		auth.EventProducerConsumer
		StartConsume(ctx context.Context) io.Closer
	}
}

func (c EventProducerConsumerContract) Test(t *testing.T) {
	t.Run(`Published event will eventually be consumed by Consumer`, func(t *testing.T) {
		pubSignupEvent := auth.SignupEvent{
			User: auth.User{
				ID:       1,
				Email:    "email",
				Username: "uname",
				Password: "pwd",
			},
		}
		now := time.Now()
		// publish event
		require.Nil(t, c.Subject.PublishUserSignupEvent(context.Background(), pubSignupEvent))

		var consumedEvent auth.ConsumedSignupEvent
		// start consumer
		c.Subject.RegisterUserSignupEventConsumer(context.Background(), func(event auth.ConsumedSignupEvent) {
			consumedEvent = event
		})
		consumer := c.Subject.StartConsume(context.Background())
		t.Cleanup(func() {
			require.Nil(t, consumer.Close())
		})

		r := testcase.Retry{Strategy: testcase.Waiter{WaitTimeout: 10 * time.Second, WaitDuration: time.Second}}
		r.Assert(t, func(tb testing.TB) {
			if consumedEvent == nil {
				tb.Fail()
				return
			}
			require.Equal(tb, pubSignupEvent, consumedEvent.SignupEvent())
			require.InDelta(t, now.Second(), consumedEvent.Timestamp().Second(), float64(5*time.Second))
		})
	})
}

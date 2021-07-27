package auth_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/adamluzsi/testcase"
	"github.com/davudsafarli/twitter/auth"
	"github.com/davudsafarli/twitter/auth/storage"
	"github.com/davudsafarli/twitter/auth/test_helpers"
	"github.com/stretchr/testify/require"
)

type FakeEventProducerConsumer struct {
	returnErr bool
}

func (m FakeEventProducerConsumer) PublishUserEvent(ctx context.Context, event auth.UserEvent) error {
	if m.returnErr {
		return fmt.Errorf("err")
	}
	return nil
}
func (m FakeEventProducerConsumer) ConsumeUserEvents(ctx context.Context, Handler func(event auth.UserEvent)) io.Closer {
	return nil
}

func TestUsecases(t *testing.T) {
	pg, err := storage.NewPostgres(test_helpers.DB_CONN_STR)
	_ = pg
	require.Nil(t, err)
	t.Run(`User can #Login after #Signup`, func(t *testing.T) {
		user := test_helpers.HopefullyUniqueUser()
		fake := FakeEventProducerConsumer{}
		uc := auth.NewUsecases(pg, fake)
		// Sign up a user
		createdUser, err := uc.SignUpUser(context.Background(), user)
		require.Nil(t, err)
		require.NotZero(t, createdUser.ID, "created user should have an ID")

		defer func() {
			require.Nil(t, pg.DeleteUser(context.Background(), createdUser.ID))
		}()

		// Login a user
		token, err := uc.Login(context.Background(), createdUser.Username, user.Password)
		require.Nil(t, err)
		require.NotEmpty(t, token)
	})

	t.Run(`When new user #Signup, it should be published and Consumer should see that event`, func(t *testing.T) {
		t.Parallel()
		user := test_helpers.HopefullyUniqueUser()
		k := test_helpers.GetEventProducerConsumer(t)
		consumed := auth.UserEvent{}
		consumer := k.ConsumeUserEvents(context.Background(), func(event auth.UserEvent) {
			consumed = event
		})
		t.Cleanup(func() {
			require.Nil(t, consumer.Close())
		})

		uc := auth.NewUsecases(pg, k)
		// Sign up a user
		createdUser, err := uc.SignUpUser(context.Background(), user)
		require.Nil(t, err)
		require.NotZero(t, createdUser.ID, "created user should have an ID")

		t.Cleanup(func() {
			require.Nil(t, pg.DeleteUser(context.Background(), createdUser.ID))
		})

		// expect the event
		buf, _ := json.Marshal(createdUser)
		expected := auth.UserEvent{
			UserID: createdUser.ID,
			Data:   buf,
		}
		r := testcase.Retry{Strategy: testcase.Waiter{WaitTimeout: 2 * time.Second, WaitDuration: time.Second / 3}}
		r.Assert(t, func(tb testing.TB) {
			require.Equal(tb, expected, consumed)
		})
	})
}

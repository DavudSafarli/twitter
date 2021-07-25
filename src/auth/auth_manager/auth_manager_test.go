package auth_manager_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/adamluzsi/testcase"
	"github.com/davudsafarli/twitter/src/auth/auth_manager"
	"github.com/davudsafarli/twitter/src/auth/storage"
	"github.com/davudsafarli/twitter/src/auth/test_helpers"
	"github.com/stretchr/testify/require"
)

type FakeEventProducerConsumer struct {
	returnErr bool
}

func (m FakeEventProducerConsumer) PublishSearchIngestEvent(ctx context.Context, event auth_manager.SearchIngestEvent) error {
	if m.returnErr {
		return fmt.Errorf("err")
	}
	return nil
}
func (m FakeEventProducerConsumer) ConsumeSearchIngestEvent(ctx context.Context, Handler func(event auth_manager.SearchIngestEvent)) io.Closer {
	return nil
}

func TestUsecases(t *testing.T) {
	pg, err := storage.NewPostgres(test_helpers.DB_CONN_STR)
	_ = pg
	require.Nil(t, err)
	t.Run(`User can #Login after #Signup`, func(t *testing.T) {
		user := test_helpers.HopefullyUniqueUser()
		fake := FakeEventProducerConsumer{}
		uc := auth_manager.NewUsecases(pg, fake)
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
		consumed := auth_manager.SearchIngestEvent{}
		consumer := k.ConsumeSearchIngestEvent(context.Background(), func(event auth_manager.SearchIngestEvent) {
			consumed = event
		})
		t.Cleanup(func() {
			require.Nil(t, consumer.Close())
		})

		uc := auth_manager.NewUsecases(pg, k)
		// Sign up a user
		createdUser, err := uc.SignUpUser(context.Background(), user)
		require.Nil(t, err)
		require.NotZero(t, createdUser.ID, "created user should have an ID")

		t.Cleanup(func() {
			require.Nil(t, pg.DeleteUser(context.Background(), createdUser.ID))
		})

		// expect the event
		buf, _ := json.Marshal(createdUser)
		expected := auth_manager.SearchIngestEvent{
			UserID: createdUser.ID,
			Data:   buf,
		}
		r := testcase.Retry{Strategy: testcase.Waiter{WaitTimeout: 2 * time.Second, WaitDuration: time.Second / 3}}
		r.Assert(t, func(tb testing.TB) {
			require.Equal(tb, expected, consumed)
		})
	})
}

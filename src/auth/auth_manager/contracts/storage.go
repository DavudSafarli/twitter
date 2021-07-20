package contracts

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/davudsafarli/twitter/src/auth"
	"github.com/davudsafarli/twitter/src/auth/auth_manager"
	"github.com/stretchr/testify/require"
)

type Subject interface {
	auth_manager.Storage
	DeleteUser(ctx context.Context, ID int) error
}

type AuthStorageContract struct {
	Subject Subject
}

func hopefullyUniqueUser() auth.User {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))

	return auth.User{
		Email:    fmt.Sprintf("Email-%X", rand.Int()),
		Username: fmt.Sprintf("Username-%X", rand.Int()),
		Password: fmt.Sprintf("Password-%X", rand.Int()),
	}
}

func (c AuthStorageContract) Test(t *testing.T) {
	t.Run(`#CreateUser + #FindUser: #FindUser returns the user if exists`, func(t *testing.T) {
		t.Parallel()
		user := hopefullyUniqueUser()

		createdUser, err := c.Subject.CreateUser(context.Background(), user)
		defer func() {
			err := c.Subject.DeleteUser(context.Background(), createdUser.ID)
			require.Nil(t, err)
		}()
		require.Nil(t, err)
		require.NotZero(t, createdUser.ID, "created user should have an ID")

		foundUser, err := c.Subject.FindUser(context.Background(), createdUser.Username)
		require.Nil(t, err)
		require.Equal(t, foundUser, createdUser)

	})

	t.Run(`#FindUser returns error if such user doesn't exist`, func(t *testing.T) {
		t.Parallel()
		foundUser, err := c.Subject.FindUser(context.Background(), `username-that-hopefully-doesnt-exist`)
		require.NotNil(t, err)
		require.Equal(t, foundUser, auth.User{})

	})
}

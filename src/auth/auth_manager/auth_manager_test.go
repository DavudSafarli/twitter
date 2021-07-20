package auth_manager

import (
	"context"
	"testing"

	"github.com/davudsafarli/twitter/src/auth"
	"github.com/davudsafarli/twitter/src/auth/storage"
	"github.com/stretchr/testify/require"
)

func TestUsecases(t *testing.T) {
	pg, err := storage.NewPostgres(auth.DB_CONN_STR)
	_ = pg
	require.Nil(t, err)
	t.Run(`Signed in user can Login`, func(t *testing.T) {
		user := auth.User{
			Email:    "davudseferli@gmail.com",
			Username: "Dave",
			Password: "pwd!@#",
		}
		uc := NewUsecases(pg)
		// Sign up a user
		createdUser, err := uc.SignUpUser(context.Background(), user)
		require.Nil(t, err)
		require.NotZero(t, createdUser.ID, "created user should have an ID")

		defer func() {
			err := pg.DeleteUser(context.Background(), createdUser.ID)
			require.Nil(t, err)
		}()

		// Login a user
		token, err := uc.Login(context.Background(), createdUser.Username, user.Password)
		require.Nil(t, err)
		require.NotEmpty(t, token)
	})

}

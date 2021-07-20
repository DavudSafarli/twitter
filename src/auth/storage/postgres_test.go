package storage

import (
	"testing"

	"github.com/davudsafarli/twitter/src/auth"
	"github.com/davudsafarli/twitter/src/auth/auth_manager/contracts"
	"github.com/stretchr/testify/require"
)

func TestPostgres(t *testing.T) {
	pg, err := NewPostgres(auth.DB_CONN_STR)
	_ = pg
	require.Nil(t, err)
	contracts.AuthStorageContract{
		Subject: pg,
	}.Test(t)
}

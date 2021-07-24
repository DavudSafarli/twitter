package storage_test

import (
	"testing"

	"github.com/davudsafarli/twitter/src/auth/auth_manager/contracts"
	"github.com/davudsafarli/twitter/src/auth/storage"
	"github.com/davudsafarli/twitter/src/auth/test_helpers"
	"github.com/stretchr/testify/require"
)

func TestPostgres(t *testing.T) {
	pg, err := storage.NewPostgres(test_helpers.DB_CONN_STR)
	_ = pg
	require.Nil(t, err)
	contracts.AuthStorageContract{
		Subject: pg,
	}.Test(t)
}

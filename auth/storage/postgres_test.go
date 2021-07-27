package storage_test

import (
	"testing"

	"github.com/davudsafarli/twitter/auth/contracts"
	"github.com/davudsafarli/twitter/auth/storage"
	"github.com/davudsafarli/twitter/auth/test_helpers"
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

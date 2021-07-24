package test_helpers

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/davudsafarli/twitter/src/auth/auth_manager"
	"github.com/davudsafarli/twitter/src/auth/event_streamer"
)

const DB_CONN_STR = "postgres://postgres:PGPWD123@localhost:7001/postgres?sslmode=disable"

var BROKERS = []string{"localhost:9094"}

func HopefullyUniqueUser() auth_manager.User {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))

	return auth_manager.User{
		Email:    fmt.Sprintf("Email-%X", rand.Int()),
		Username: fmt.Sprintf("Username-%X", rand.Int()),
		Password: fmt.Sprintf("Password-%X", rand.Int()),
	}
}

func GetEventProducerConsumer() auth_manager.EventProducerConsumer {
	k := event_streamer.NewKafka(BROKERS)
	return k
}

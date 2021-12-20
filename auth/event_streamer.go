package auth

import (
	"context"
	"time"
)

type SignupEvent struct {
	// all user fields are necessary for the event right now.
	// If User struct grows to have fields we don't want to send as an event, then copy the fields here
	User
}

type EventProducerConsumer interface {
	PublishUserSignupEvent(ctx context.Context, event SignupEvent) error
	RegisterUserSignupEventConsumer(ctx context.Context, Handler func(event ConsumedSignupEvent))
}

type ConsumedSignupEvent interface {
	Timestamp() time.Time
	SignupEvent() SignupEvent
}

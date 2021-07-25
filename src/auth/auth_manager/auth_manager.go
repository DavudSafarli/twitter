package auth_manager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type Storage interface {
	CreateUser(ctx context.Context, u User) (User, error)
	FindUser(ctx context.Context, usnm string) (User, error)
}

type SearchIngestEvent struct {
	UserID int
	Data   []byte
}

type EventProducerConsumer interface {
	PublishSearchIngestEvent(ctx context.Context, event SearchIngestEvent) error
	ConsumeSearchIngestEvent(ctx context.Context, Handler func(event SearchIngestEvent)) io.Closer
}

type Usecases struct {
	Storage  Storage
	Publiser EventProducerConsumer
}

func NewUsecases(s Storage, publisher EventProducerConsumer) Usecases {
	return Usecases{
		Storage:  s,
		Publiser: publisher,
	}
}

// TODO: Read from config
const JWT_SECRET = "jwt_secret"

// SignUpUser registers a new user if the username and email don't exist already.
// It hashes the password before saving.
// It also publishes 2 events. One for `SearchIngestor`, and one for `SocialGraphBuilder`.
func (c Usecases) SignUpUser(ctx context.Context, user User) (User, error) {
	hashedPwd, err := hashPassword(user.Password)
	if err != nil {
		return User{}, err
	}
	user.Password = hashedPwd
	user, err = c.Storage.CreateUser(ctx, user)
	if err != nil {
		return User{}, err
	}
	buf, err := json.Marshal(user)
	if err != nil {
		return User{}, err
	}
	err = c.Publiser.PublishSearchIngestEvent(ctx, SearchIngestEvent{
		UserID: user.ID,
		Data:   buf,
	})

	return user, err
}

// Login creates and retunrs a token for an existing user.
func (c Usecases) Login(ctx context.Context, usnm, pwd string) (token string, err error) {
	user, err := c.Storage.FindUser(ctx, usnm)
	// TODO: return user doesn't exist error
	if err != nil {
		return "", err
	}

	correct := checkPasswordAndHashEquality(pwd, user.Password)
	if !correct {
		return "", errors.New("wrong password")
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"ID":  fmt.Sprint(user.ID),
		"nbf": time.Date(2015, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
	})

	token, err = jwtToken.SignedString([]byte(JWT_SECRET))

	return token, err
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordAndHashEquality(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

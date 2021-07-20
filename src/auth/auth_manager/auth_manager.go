package auth_manager

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/davudsafarli/twitter/src/auth"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type Storage interface {
	CreateUser(ctx context.Context, u auth.User) (auth.User, error)
	FindUser(ctx context.Context, usnm string) (auth.User, error)
}

type Usecases struct {
	Storage Storage
}

func NewUsecases(s Storage) Usecases {
	return Usecases{
		Storage: s,
	}
}

const JWT_SECRET = "jwt_secret"

// SignUpUser registers a new user if the username and email don't exist already.
// It hashes the password before saving.
// It also publishes 2 events. One for `SearchIngestor`, and one for `SocialGraphBuilder`.
func (c Usecases) SignUpUser(ctx context.Context, user auth.User) (auth.User, error) {
	hashedPwd, err := hashPassword(user.Password)
	if err != nil {
		return auth.User{}, err
	}
	user.Password = hashedPwd
	return c.Storage.CreateUser(ctx, user)
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

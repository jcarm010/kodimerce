package entities

import (
	"errors"
	"github.com/jcarm010/kodimerce/datastore"
	"github.com/satori/go.uuid"
	"golang.org/x/net/context"
)

const (
	EntityUser        = "user"
	EntityUserSession = "user_session"
)

var (
	ErrUserAlreadyExists = errors.New("User already exists.")
)

type User struct {
	Email           string `json:"email" datastore:"-"`
	PasswordHash    string `json:"password_hash" datastore:"password_hash,noindex"`
	UserType        string `json:"user_type" datastore:"user_type"`
	LastVisitedPath string `json:"last_visited_path" datastore:"last_visited_path"`
}

func NewUser(email string) *User {
	return &User{
		Email:    email,
		UserType: "regular",
	}
}

type UserSession struct {
	Email        string `json:"email" datastore:"email"`
	SessionToken string `json:"session_token" datastore:"-"`
}

func NewUserSession(sessionToken string, email string) *UserSession {
	return &UserSession{
		SessionToken: sessionToken,
		Email:        email,
	}
}

func CreateUser(ctx context.Context, user *User) error {
	key := datastore.NewKey(ctx, EntityUser, user.Email, 0, nil)
	err := datastore.RunInTransaction(ctx, func(transaction *datastore.Transaction) error {
		u := &User{}
		err := transaction.Get(key, u)
		if err != nil && err != datastore.ErrNoSuchEntity {
			return err
		} else if err == nil {
			return ErrUserAlreadyExists
		}

		_, err = transaction.Put(key, user)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func UpdateUser(ctx context.Context, user *User) error {
	key := datastore.NewKey(ctx, EntityUser, user.Email, 0, nil)
	_, err := datastore.Put(ctx, key, user)
	return err
}

func GetUser(ctx context.Context, email string) (*User, error) {
	key := datastore.NewKey(ctx, EntityUser, email, 0, nil)
	u := &User{}
	err := datastore.Get(ctx, key, u)
	if err != nil {
		return nil, err
	}

	u.Email = email
	return u, nil
}

func CreateUserSession(ctx context.Context, email string) (*UserSession, error) {
	nUdid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	userSession := NewUserSession(nUdid.String(), email)
	key := datastore.NewKey(ctx, EntityUserSession, userSession.SessionToken, 0, nil)
	_, err = datastore.Put(ctx, key, userSession)
	if err != nil {
		return nil, err
	}

	return userSession, nil
}

func GetUserSession(ctx context.Context, sessionToken string) (*UserSession, error) {
	key := datastore.NewKey(ctx, EntityUserSession, sessionToken, 0, nil)
	userSession := &UserSession{}
	err := datastore.Get(ctx, key, userSession)
	if err != nil {
		return nil, err
	}

	userSession.SessionToken = sessionToken
	return userSession, nil
}

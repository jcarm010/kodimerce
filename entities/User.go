package entities

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"errors"
	"github.com/satori/go.uuid"
)

const(
	ENTITY_USER = "user"
	ENTITY_USER_SESSION = "user_session"
)

var (
	ErrUserAlreadyExists = errors.New("User already exists.")
)

type User struct {
	Email string `json:"email" datastore:"-"`
	PasswordHash string `json:"password_hash" datastore:"password_hash,noindex"`
	UserType string `json:"user_type" datastore:"user_type"`
	LastVisitedPath string `json:"last_visited_path" datastore:"last_visited_path"`
}

func NewUser(email string) *User {
	return &User{
		Email: email,
		UserType: "regular",
	}
}

type UserSession struct {
	Email string `json:"email" datastore:"email"`
	SessionToken string `json:"session_token" datastore:"-"`
}

func NewUserSession(sessionToken string, email string) *UserSession {
	return &UserSession{
		SessionToken: sessionToken,
		Email: email,
	}
}

func CreateUser(ctx context.Context, user *User) error {
	key := datastore.NewKey(ctx, ENTITY_USER, user.Email, 0, nil)
	err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		u := &User{}
		err := datastore.Get(ctx, key, u)
		if err != nil && err != datastore.ErrNoSuchEntity {
			return err
		}else if err == nil {
			return ErrUserAlreadyExists
		}

		_, err = datastore.Put(ctx, key, user)
		if err != nil {
			return err
		}

		return nil
	}, nil)

	if err != nil {
		return err
	}

	return nil
}

func UpdateUser(ctx context.Context, user *User) error {
	key := datastore.NewKey(ctx, ENTITY_USER, user.Email, 0, nil)
	_, err := datastore.Put(ctx, key, user)
	return err
}

func GetUser(ctx context.Context, email string) (*User, error) {
	key := datastore.NewKey(ctx, ENTITY_USER, email, 0, nil)
	u := &User{}
	err := datastore.Get(ctx, key, u)
	if err != nil {
		return nil, err
	}

	u.Email = email
	return u, nil
}

func CreateUserSession(ctx context.Context, email string) (*UserSession, error) {
	userSession := NewUserSession(uuid.NewV4().String(), email)
	key := datastore.NewKey(ctx, ENTITY_USER_SESSION, userSession.SessionToken, 0, nil)
	_, err := datastore.Put(ctx, key, userSession)
	if err != nil{
		return nil, err
	}

	return userSession, nil
}

func GetUserSession(ctx context.Context, sessionToken string) (*UserSession, error) {
	key := datastore.NewKey(ctx, ENTITY_USER_SESSION, sessionToken, 0, nil)
	userSession := &UserSession{}
	err := datastore.Get(ctx, key, userSession)
	if err != nil{
		return nil, err
	}

	userSession.SessionToken = sessionToken
	return userSession, nil
}
package entities

import (
	"github.com/dustin/gojson"
	"log"
)

const (
	USER_TYPE_OWNER = "owner"
)

type User struct {
	Email string
	FirstName string `datastore:",noindex"`
	LastName string `datastore:",noindex"`
	Role string `datastore:",noindex"`
	PasswordHash string `datastore:",noindex" json:"-"`
}

func NewUser(email string, firstName string, lastName string, role string, passwordHash string) *User {
	return &User{
		Email: email,
		FirstName: firstName,
		LastName: lastName,
		Role: role,
		PasswordHash: passwordHash,
	}
}

func (c *User) Json() string {
	json, _ := json.Marshal(c)
	jsonStr := string(json)
	log.Printf("Json String: %+v", jsonStr)
	return jsonStr
}
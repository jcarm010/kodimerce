package entities

type User struct {
	Email string
	FirstName string
	LastName string
	PasswordHash string
}

func NewUser(email string, firstName string, lastName string, passwordHash string) *User {
	return &User{
		Email: email,
		FirstName: firstName,
		LastName: lastName,
		PasswordHash: passwordHash,
	}
}

package entities

import (
	"testing"
	"reflect"
)

func TestNewUser(t *testing.T) {
	want := &User{
		Email: TEST_USER_EMAIL,
		FirstName: TEST_USER_FIRST_NAME,
		LastName: TEST_USER_LAST_NAME,
		PasswordHash: TEST_PASSWORD_HASH,
	}
	got := NewUser(TEST_USER_EMAIL, TEST_USER_FIRST_NAME, TEST_USER_LAST_NAME, TEST_PASSWORD_HASH)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Configurations are not equal.\nNeed: %+v\nGot: %+v", want, got)
	}
}

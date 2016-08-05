package entities

import (
	"reflect"
	"testing"
	"google.golang.org/appengine/aetest"
	"golang.org/x/crypto/bcrypt"
)

const TEST_USER_EMAIL = "test@kodimerce.com"
const TEST_USER_FIRST_NAME = "John"
const TEST_USER_LAST_NAME = "Doe"
const TEST_USER_PASSWORD = "P@ssword"
const TEST_COMPANY_NAME = "Test Company"
const TEST_COMPANY_ADDRESS = "8888 SW 87 Ave"
const TEST_COMPANY_EMAIL = "contact@kodimerce.com"
const TEST_COMPANY_PHONE = "8888888888"

var TEST_PASSWORD_BYTES, _ = bcrypt.GenerateFromPassword([]byte(TEST_USER_PASSWORD), bcrypt.DefaultCost)
var TEST_PASSWORD_HASH = string(TEST_PASSWORD_BYTES)

func TestNewServerConfig(t *testing.T) {
	want := ServerConfig{
		ConfigKey: SERVER_CONFIG_KEY,
		CompanyName: TEST_COMPANY_NAME,
		CompanyAddress: TEST_COMPANY_ADDRESS,
		CompanyEmail: TEST_COMPANY_EMAIL,
		CompanyPhone: TEST_COMPANY_PHONE,
	}

	config := NewServerConfig(TEST_COMPANY_NAME, TEST_COMPANY_ADDRESS, TEST_COMPANY_EMAIL, TEST_COMPANY_PHONE)
	if !reflect.DeepEqual(config, want) {
		t.Errorf("Configurations are not equal.\nNeed: %+v\nGot: %+v", want, config)
	}
}

func TestGetServerConfig(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	config, err := GetServerConfig(ctx)
	if err != ENTITY_NOT_FOUND_ERROR {
		t.Errorf("Server confuguration should not have been found but instead got config[%+v] err[%+v]", config, err)
	}
}

func TestSetServerConfig(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	config, err := GetServerConfig(ctx)
	if err != ENTITY_NOT_FOUND_ERROR {
		t.Errorf("Server confuguration should not have been found but instead got config[%+v] err[%+v]", config, err)
		return
	}
	want := NewServerConfig(TEST_COMPANY_NAME, TEST_COMPANY_ADDRESS, TEST_COMPANY_EMAIL, TEST_COMPANY_PHONE)
	err = SetServerConfig(ctx, want)
	if err != nil {
		t.Errorf("Error setting server config: %+v", err)
		return
	}
	config, err = GetServerConfig(ctx)
	if err != nil {
		t.Errorf("Error getting created configuration: %+v", err)
		return
	}
	if !reflect.DeepEqual(config, want) {
		t.Errorf("Configurations are not equal.\nNeed: %+v\nGot: %+v", want, config)
	}
}

func TestGetUser(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	user := NewUser(TEST_USER_EMAIL, TEST_USER_FIRST_NAME, TEST_USER_LAST_NAME, USER_TYPE_OWNER, TEST_PASSWORD_HASH)
	err = CreateUser(ctx, user)
	if err != nil {
		t.Errorf("Failed to create user %+v", err)
		return
	}
	userFetched, err := GetUser(ctx, TEST_USER_EMAIL)
	if err != nil {
		t.Errorf("Failed to fetch user: %+v", err)
		return
	}
	if !reflect.DeepEqual(user, userFetched) {
		t.Errorf("Configurations are not equal.\nNeed: %+v\nGot: %+v", user, userFetched)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(userFetched.PasswordHash), []byte(TEST_USER_PASSWORD))
	if err != nil {
		t.Errorf("Passwords do not match: %+v", err)
		return
	}
}

func TestCreateUser(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	user := NewUser(TEST_USER_EMAIL, TEST_USER_FIRST_NAME, TEST_USER_LAST_NAME, USER_TYPE_OWNER, TEST_PASSWORD_HASH)
	err = CreateUser(ctx, user)
	if err != nil {
		t.Errorf("Failed to create user %+v", err)
		return
	}
}
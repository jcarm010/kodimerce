package entities

import (
	"reflect"
	"testing"
	"google.golang.org/appengine/aetest"
)

const companyName = "Test Company"
const companyAddress = "8888 SW 87 Ave"
const companyEmail = "contact@kodimerce.com"
const companyPhone = "8888888888"

func TestNewServerConfig(t *testing.T) {
	want := ServerConfig{
		ConfigKey: SERVER_CONFIG_KEY,
		CompanyName: companyName,
		CompanyAddress: companyAddress,
		CompanyEmail: companyEmail,
		CompanyPhone: companyPhone,
	}

	config := NewServerConfig(companyName, companyAddress, companyEmail, companyPhone)
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
	want := NewServerConfig(companyName, companyAddress, companyEmail, companyPhone)
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
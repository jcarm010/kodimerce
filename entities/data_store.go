package entities

import (
	"google.golang.org/appengine/datastore"
	"golang.org/x/net/context"
	"errors"
	"google.golang.org/appengine/memcache"
	"google.golang.org/appengine/log"
	"github.com/dustin/gojson"
)

const SERVER_CONFIG_KEY = "sck"
var ENTITY_NOT_FOUND_ERROR = errors.New("Entity not found")

type ServerConfig struct {
	ConfigKey string
	CompanyName string `datastore:",noindex"`
	CompanyAddress string `datastore:",noindex"`
	CompanyEmail string `datastore:",noindex"`
	CompanyPhone string `datastore:",noindex"`
}

func NewServerConfig(companyName string, companyAddress string, companyEmail string, companyPhone string) ServerConfig {
	return ServerConfig{
		ConfigKey: SERVER_CONFIG_KEY,
		CompanyName: companyName,
		CompanyAddress: companyAddress,
		CompanyEmail: companyEmail,
		CompanyPhone: companyPhone,
	}
}

func SetServerConfig(ctx context.Context, config ServerConfig) error {
	log.Infof(ctx, "Storing config: %+v", config)
	key := datastore.NewIncompleteKey(ctx, "ServerConfig", nil)
	_, err := datastore.Put(ctx, key, &config)
	if err != nil{
		return err
	}

	jsonData, err := json.Marshal(config)
	if err != nil {
		log.Errorf(ctx, "Failed to json encode configuration for memcache: %+v", err)
		return nil
	}

	item := &memcache.Item{
		Key:   SERVER_CONFIG_KEY,
		Value: jsonData,
	}

	if err := memcache.Set(ctx, item); err != nil {
		log.Errorf(ctx, "Error setting memcache config: %v", err)
	}

	return nil
}

func GetServerConfig(ctx context.Context) (ServerConfig, error) {
	item, err := memcache.Get(ctx, SERVER_CONFIG_KEY)
	if err == nil{
		c := ServerConfig{}
		err := json.Unmarshal(item.Value, &c)
		if err != nil {
			log.Errorf(ctx, "Error parsing memcache server config: %+v", err)
		}
		log.Infof(ctx, "Memcache hit: %+v", c)
		return c, nil
	}else if(err != memcache.ErrCacheMiss){
		log.Errorf(ctx, "Error fetching server config from memcache: %+v", err)
	}else {
		log.Infof(ctx, "Server config not in cache")
	}

	var config []ServerConfig
	query := datastore.NewQuery("ServerConfig").Filter("ConfigKey=", SERVER_CONFIG_KEY).Limit(1)
	_, err = query.GetAll(ctx, &config)
	if err != nil{
		return ServerConfig{}, err
	}

	if len(config) == 0{
		return ServerConfig{}, ENTITY_NOT_FOUND_ERROR
	}

	jsonData, err := json.Marshal(config[0])
	if err != nil {
		log.Errorf(ctx, "Failed to json encode configuration for memcache: %+v", err)
		return config[0], nil
	}

	item = &memcache.Item{
		Key:   SERVER_CONFIG_KEY,
		Value: jsonData,
	}

	if err := memcache.Set(ctx, item); err != nil {
		log.Errorf(ctx, "Error setting memcache config: %v", err)
	}

	return config[0], nil
}

func CreateUser(ctx context.Context, user *User) error {
	key := datastore.NewKey(ctx, "User", user.Email, 0, nil)
	key, err := datastore.Put(ctx, key, user)
	if err != nil {
		return err
	}
	return nil
}

func GetUser(ctx context.Context, email string) (*User, error) {
	user := &User{}
	err := datastore.Get(ctx, datastore.NewKey(ctx, "User", email, 0, nil), user)
	if err != nil {
		return nil, err
	}
	return user, nil
}
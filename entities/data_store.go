package entities

import (
	"google.golang.org/appengine/datastore"
	"golang.org/x/net/context"
	"errors"
	"google.golang.org/appengine/memcache"
	"google.golang.org/appengine/log"
	"github.com/dustin/gojson"
	"time"
)

const (
	CONFIG_KEY_SERVER = "server_config"
	CONFIG_KEY_COMPANY = "company_config"
)

var ENTITY_NOT_FOUND_ERROR = errors.New("Entity not found")

const (
	ENTITY_SERVER_CONFIG = "ServerConfig"
	ENTITY_USER = "User"
	ENTITY_SESSION_TOKEN = "SessionToken"
)

type SessionToken struct {
	Email string
	Token string
	Created time.Time
}

func NewSessionToken(email string, token string) *SessionToken{
	return &SessionToken{
		Email: email,
		Token: token,
		Created: time.Now(),
	}
}

func SetCompanyConfig(ctx context.Context, config CompanyConfig) error {
	log.Infof(ctx, "Storing config: %+v", config)
	key := datastore.NewKey(ctx, ENTITY_SERVER_CONFIG, CONFIG_KEY_COMPANY, 0, nil)
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
		Key:   CONFIG_KEY_COMPANY,
		Value: jsonData,
	}

	if err := memcache.Set(ctx, item); err != nil {
		log.Errorf(ctx, "Error setting memcache config: %v", err)
	}

	return nil
}

func GetCompanyConfig(ctx context.Context) (CompanyConfig, error) {
	item, err := memcache.Get(ctx, CONFIG_KEY_COMPANY)
	if err == nil{
		c := CompanyConfig{}
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

	var config []CompanyConfig
	query := datastore.NewQuery(ENTITY_SERVER_CONFIG).Filter("ConfigKey=", CONFIG_KEY_COMPANY).Limit(1)
	_, err = query.GetAll(ctx, &config)
	if err != nil{
		return CompanyConfig{}, err
	}

	if len(config) == 0{
		return CompanyConfig{}, ENTITY_NOT_FOUND_ERROR
	}

	jsonData, err := json.Marshal(config[0])
	if err != nil {
		log.Errorf(ctx, "Failed to json encode configuration for memcache: %+v", err)
		return config[0], nil
	}

	item = &memcache.Item{
		Key:   CONFIG_KEY_COMPANY,
		Value: jsonData,
	}

	if err := memcache.Set(ctx, item); err != nil {
		log.Errorf(ctx, "Error setting memcache config: %v", err)
	}

	return config[0], nil
}

func SetServerConfig(ctx context.Context, config ServerConfig) error {
	log.Infof(ctx, "Storing config: %+v", config)
	key := datastore.NewKey(ctx, ENTITY_SERVER_CONFIG, CONFIG_KEY_SERVER, 0, nil)
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
		Key:   CONFIG_KEY_SERVER,
		Value: jsonData,
	}

	if err := memcache.Set(ctx, item); err != nil {
		log.Errorf(ctx, "Error setting memcache config: %v", err)
	}

	return nil
}

func GetServerConfig(ctx context.Context) (ServerConfig, error) {
	item, err := memcache.Get(ctx, CONFIG_KEY_SERVER)
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
	query := datastore.NewQuery(ENTITY_SERVER_CONFIG).Filter("ConfigKey=", CONFIG_KEY_SERVER).Limit(1)
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
		Key:   CONFIG_KEY_SERVER,
		Value: jsonData,
	}

	if err := memcache.Set(ctx, item); err != nil {
		log.Errorf(ctx, "Error setting memcache config: %v", err)
	}

	return config[0], nil
}

func CreateUser(ctx context.Context, user *User) error {
	key := datastore.NewKey(ctx, ENTITY_USER, user.Email, 0, nil)
	key, err := datastore.Put(ctx, key, user)
	if err != nil {
		return err
	}

	return nil
}

func GetUser(ctx context.Context, email string) (*User, error) {
	user := &User{}
	err := datastore.Get(ctx, datastore.NewKey(ctx, ENTITY_USER, email, 0, nil), user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func StoreSessionToken(ctx context.Context, sessionToken *SessionToken) error {
	key := datastore.NewKey(ctx, ENTITY_SESSION_TOKEN, sessionToken.Token, 0, nil)
	_, err := datastore.Put(ctx, key, sessionToken)
	if err != nil {
		return err
	}

	return nil
}

func GetSessionToken(ctx context.Context, token string) (*SessionToken,error) {
	sessionToken := &SessionToken{}
	err := datastore.Get(ctx, datastore.NewKey(ctx, ENTITY_SESSION_TOKEN, token, 0, nil), sessionToken)
	if err != nil {
		return nil, err
	}

	return sessionToken, nil
}
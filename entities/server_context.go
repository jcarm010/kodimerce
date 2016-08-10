package entities

import (
	"github.com/gocraft/web"
	"fmt"
	"google.golang.org/appengine"
	"encoding/json"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	"net/http"
)

type ServerContext struct {
	Context       context.Context
	w             web.ResponseWriter
	r             *web.Request
	ServerConfig  ServerConfig
	CompanyConfig CompanyConfig
	SessionToken  *SessionToken
	User          *User
}

func (c *ServerContext) SetServerConfiguration(w web.ResponseWriter, r *web.Request, next web.NextMiddlewareFunc) {
	c.Context = appengine.NewContext(r.Request)
	c.w = w
	c.r = r
	companyConfig, err := GetCompanyConfig(c.Context)
	if(err == ENTITY_NOT_FOUND_ERROR){
		if(r.RequestURI != "/init"){
			http.Redirect(w, r.Request, "/init", http.StatusTemporaryRedirect)
			return
		}
	} else if err != nil {
		log.Errorf(c.Context, "Error retrieving server config: %+v", companyConfig)
		c.ServeJson(500, "")
		return
	} else if r.RequestURI == "/init" || r.RequestURI == "/init/"{
		http.Redirect(w, r.Request, "/", http.StatusMovedPermanently)
		return
	}

	log.Infof(c.Context, "Company Config: %+v", companyConfig)
	c.CompanyConfig = companyConfig

	serverConfig, err := GetServerConfig(c.Context)
	if(err == ENTITY_NOT_FOUND_ERROR){
		serverConfig = DefaultServerConfig()
		err = SetServerConfig(c.Context, serverConfig)
		if err != nil {
			log.Errorf(c.Context, "Failed to save server config: %+v", err)
		}
	} else if err != nil {
		log.Errorf(c.Context, "Error retrieving server config: %+v", companyConfig)
		serverConfig = DefaultServerConfig()
	}

	log.Infof(c.Context, "Server Config: %+v", serverConfig)
	c.ServerConfig = serverConfig
	next(w, r)
}

func (c *ServerContext) ValidateAdminUser(w web.ResponseWriter, r *web.Request, next web.NextMiddlewareFunc) {
	err := c.SetUserContext()
	if err != nil {
		c.ServeJson(http.StatusUnauthorized, "Not Authorizard")
		return
	}
	next(w,r)
}

func (c *ServerContext) SetUserContext() error {
	cookie, err := c.r.Cookie("session")
	if err != nil {
		return err
	}

	token := cookie.Value
	sessionToken, err := GetSessionToken(c.Context, token)
	if err != nil {
		return err
	}

	c.SessionToken = sessionToken

	user, err := GetUser(c.Context, sessionToken.Email)
	if err != nil {
		return err
	}

	c.User = user

	return nil
}

func (c *ServerContext) ServeJson(status int, value interface{}){
	c.w.Header().Add("Content-Type", "application/json")
	c.w.WriteHeader(status)
	bts, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(c.w, "%s", bts)
}

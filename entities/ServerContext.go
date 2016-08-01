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
	Context context.Context
	w web.ResponseWriter
	r *web.Request
	Config ServerConfig
}

func (c *ServerContext) SetServerConfiguration(w web.ResponseWriter, r *web.Request, next web.NextMiddlewareFunc){
	c.Context = appengine.NewContext(r.Request)
	c.w = w
	c.r = r
	serverConfig, err := GetServerConfig(c.Context)
	if(err == ENTITY_NOT_FOUND_ERROR){
		if(r.RequestURI != "/init"){
			http.Redirect(w, r.Request, "/init", http.StatusTemporaryRedirect)
			return
		}
	} else if err != nil {
		log.Errorf(c.Context, "Error retrieving server config: %+v", serverConfig)
		c.ServeJson(500, "")
		return
	} else if r.RequestURI == "/init" || r.RequestURI == "/init/"{
		http.Redirect(w, r.Request, "/", http.StatusMovedPermanently)
		return
	}

	log.Infof(c.Context, "Server Config: %+v", serverConfig)
	c.Config = serverConfig

	next(w, r)
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

package server

import (
	"github.com/gocraft/web"
	"net/http"
	"login"
)

type ServerContext struct {

}

func init() {
	router := web.New(ServerContext{}).
		Middleware(web.LoggerMiddleware).
		Middleware(web.ShowErrorsMiddleware).
		Get("/login", login.Login)

	http.Handle("/", router)
}
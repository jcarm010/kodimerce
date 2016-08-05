package server

import (
	"github.com/gocraft/web"
	"net/http"
	"login"
	"entities"
	"initial_config"
	"admin"
)

func init() {
	router := web.New(entities.ServerContext{}).
		Middleware(web.LoggerMiddleware).
		Middleware(web.ShowErrorsMiddleware).
		Get("/init", initial_config.ServerInit).
		Post("/init", initial_config.SetupServerInit).
		Middleware((*entities.ServerContext).SetServerConfiguration).
		Get("/login", login.LoginView).
		Post("/login", login.Login).
		Get("/login/landing", login.UserLanding)

	router.Subrouter(entities.ServerContext{}, "/admin").
		Middleware((*entities.ServerContext).ValidateAdminUser).
		Get("/", admin.AdminView).
		Get("/:", admin.AdminView)

	http.Handle("/", router)
}
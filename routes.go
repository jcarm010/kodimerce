package server

import (
	"github.com/gocraft/web"
	"net/http"
	"entities"
	"km"
)

func init() {
	router := web.New(entities.ServerContext{}).
		Middleware(web.LoggerMiddleware).
		Middleware(web.ShowErrorsMiddleware).
		Get("/init", km.ServerInitView).
		Post("/init", km.SetupServerInit).
		Middleware((*entities.ServerContext).SetServerConfiguration).
		Get("/login", km.LoginView).
		Post("/login", km.Login).
		Get("/login/landing", km.UserLanding)

	router.Subrouter(entities.ServerContext{}, "/admin").
		Middleware((*entities.ServerContext).ValidateAdminUser).
		Put("/company", km.UpdateCompanyDetails).
		Get("/", km.AdminView).
		Get("/:", km.AdminView)

	http.Handle("/", router)
}
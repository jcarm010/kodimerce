package server

import (
	"github.com/gocraft/web"
	"net/http"
	"km"
	"views"
)

func init() {
	router := web.New(km.ServerContext{}).
		Middleware(web.LoggerMiddleware).
		Middleware((*km.ServerContext).InitServerContext)

	router.Subrouter(km.AdminContext{}, "/admin").
		Middleware((*km.AdminContext).Auth).
		Get("/km/product", (*km.AdminContext).GetProducts).
		Post("/km/product", (*km.AdminContext).CreateProduct).
		Put("/km/product", (*km.AdminContext).UpdateProduct).
		Get("/", views.AdminView).
		Get("/:page", views.AdminView).
		Get("/:page/:subpage", views.AdminView)

	http.Handle("/", router)
}
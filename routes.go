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
		Middleware((*km.ServerContext).InitServerContext).
		Get("/", views.HomeView).
		Get("/register", views.RegisterView).
		Post("/register", (*km.ServerContext).RegisterUser).
		Get("/login", views.LoginView).
		Post("/login", (*km.ServerContext).LoginUser)

	router.Subrouter(km.AdminContext{}, "/admin").
		Middleware((*km.AdminContext).Auth).
		Get("/km/product", (*km.AdminContext).GetProducts).
		Post("/km/product", (*km.AdminContext).CreateProduct).
		Put("/km/product", (*km.AdminContext).UpdateProduct).
		Get("/km/category", (*km.AdminContext).GetCategory).
		Post("/km/category", (*km.AdminContext).CreateCategory).
		Put("/km/category", (*km.AdminContext).UpdateCategory).
		Get("/km/category_products", (*km.AdminContext).GetCategoryProduct).
		Post("/km/category_products", (*km.AdminContext).SetCategoryProducts).
		Delete("/km/category_products", (*km.AdminContext).UnsetCategoryProducts).
		Get("/", views.AdminView).
		Get("/:page", views.AdminView).
		Get("/:page/:subpage", views.AdminView)

	http.Handle("/", router)
}
package server

import (
	"github.com/gocraft/web"
	"net/http"
	"github.com/jcarm010/kodimerce/km"
	"github.com/jcarm010/kodimerce/views"
)

func init() {
	router := web.New(km.ServerContext{}).
		Middleware(web.LoggerMiddleware).
		Middleware((*km.ServerContext).InitServerContext).
		Get("/", views.HomeView).
		Get("/contact", views.ContactView).
		Post("/contact", (*km.ServerContext).PostContactMessage).
		Get("/product", views.ProductView).
		Get("/product/:productId", views.ProductView).
		Get("/store", views.StoreView).
		Get("/store/:category", views.StoreView).
		Get("/register", views.RegisterView).
		Post("/register", (*km.ServerContext).RegisterUser).
		Get("/login", views.LoginView).
		Post("/login", (*km.ServerContext).LoginUser).
		Get("/cart", views.CartView).
		Get("/checkout", views.RenderCheckoutView).
		Get("/checkout/:step", views.RenderCheckoutView).
		Get("/order", views.OrderReviewView).
		Post("/order/address/verify", (*km.ServerContext).CheckOrderAddress).
		Post("/order", (*km.ServerContext).CreateOrder).
		Put("/order", (*km.ServerContext).UpdateOrder).
		Get("/paypal/payment", (*km.ServerContext).CreatePaypalPayment).
		Post("/paypal/payment", (*km.ServerContext).ExecutePaypalPayment).
		Get("/gallery/upload", (*km.ServerContext).GetGalleryUpload)


	router.Subrouter(km.ServerContext{}, "/api").
		Get("/product", (*km.ServerContext).GetProducts)

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
		Get("/gallery/upload", (*km.AdminContext).GetGalleryUploads).
		Post("/gallery/upload", (*km.AdminContext).PostGalleryUpload).
		Delete("/gallery/upload", (*km.AdminContext).DeleteGalleryUpload).
		Get("/gallery/upload/url", (*km.AdminContext).GetGalleryUploadUrl).
		Get("/order", (*km.AdminContext).GetOrders).
		Get("/", views.AdminView).
		/* Write new admin endpoints above. These two need to be the last admin endpoints. */
		Get("/:page", views.AdminView).
		Get("/:page/:subpage", views.AdminView)

	http.Handle("/", router)
}
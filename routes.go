package server

import (
	"github.com/gocraft/web"
	"github.com/jcarm010/kodimerce/km"
	"net/http"
	"github.com/jcarm010/kodimerce/views"
)

func init() {
	router := web.New(km.ServerContext{}).
		Middleware(web.LoggerMiddleware).
		Middleware((*km.ServerContext).InitServerContext).
		Middleware((*km.ServerContext).SetRedirects).
		Middleware((*km.ServerContext).SetCORS)

	router = router.Middleware((*km.ServerContext).RedirectWWW)

	router = router.Get("/robots.txt", views.RobotsFile).
		Get("/", views.HomeView).
		Get("/contact", views.ContactView).
		Get("/referrals", views.ReferralsView).
		Post("/contact", (*km.ServerContext).PostContactMessage).
		Get("/product", views.ProductView).
		Get("/product/:productId", views.ProductView).
		Get("/store", views.StoreView).
		Get("/store/:category", views.StoreView).
		Get("/gallery", views.GalleriesView).
		Get("/gallery/:galleryPath", views.GalleryView).
		Get("/register", views.RegisterView).
		Post("/register", (*km.ServerContext).RegisterUser).
		Get("/login", views.LoginView).
		Post("/login", (*km.ServerContext).LoginUser).
		Get("/cart", views.CartView).
		Get("/checkout", views.RenderCheckoutView).
		Get("/checkout/:step", views.RenderCheckoutView).
		Get("/thank-you", views.ThankYouView).
		Get("/order", views.OrderReviewView).
		Post("/order/address/verify", (*km.ServerContext).CheckOrderAddress).
		Post("/order", (*km.ServerContext).CreateOrder).
		Put("/order", (*km.ServerContext).UpdateOrder).
		Get("/paypal/payment", (*km.ServerContext).CreatePaypalPayment).
		Post("/paypal/payment", (*km.ServerContext).ExecutePaypalPayment).
		Get("/gallery/upload", (*km.ServerContext).GetGalleryUpload).
		Get("/gallery/upload/name/:name", (*km.ServerContext).GetGalleryUploadByName).
		Get("/gallery/upload/:key", (*km.ServerContext).GetGalleryUpload).
		Get("/sitemap.xml", (*km.ServerContext).GetSiteMap).
		Get("/blog", views.BlogView).
		Get("/blog/rss", views.GetBlogRss).
		Get("/amp/:path", views.GetAmpDynamicPage).
		Get("/:*", views.GetDynamicPage)

	router.Subrouter(km.ServerContext{}, "/api").
		Get("/product", (*km.ServerContext).GetProducts)

	router.Subrouter(km.AdminContext{}, "/admin").
		Middleware((*km.AdminContext).Auth).
		Post("/km/last/visited/path", (*km.AdminContext).SaveLastVisitedPath).
		Get("/km/post", (*km.AdminContext).GetPosts).
		Post("/km/post", (*km.AdminContext).CreatePost).
		Put("/km/post", (*km.AdminContext).UpdatePost).
		Get("/km/product", (*km.AdminContext).GetProducts).
		Post("/km/product", (*km.AdminContext).CreateProduct).
		Put("/km/product", (*km.AdminContext).UpdateProduct).
		Get("/km/category", (*km.AdminContext).GetCategory).
		Post("/km/category", (*km.AdminContext).CreateCategory).
		Put("/km/category", (*km.AdminContext).UpdateCategory).
		Post("/km/page", (*km.AdminContext).CreatePage).
		Get("/km/page", (*km.AdminContext).GetPages).
		Put("/km/page", (*km.AdminContext).UpdatePage).
		Get("/km/category_products", (*km.AdminContext).GetCategoryProduct).
		Post("/km/category_products", (*km.AdminContext).SetCategoryProducts).
		Delete("/km/category_products", (*km.AdminContext).UnsetCategoryProducts).
		Get("/km/gallery", (*km.AdminContext).GetGalleries).
		Post("/km/gallery", (*km.AdminContext).CreateGallery).
		Put("/km/gallery", (*km.AdminContext).UpdateGallery).
		Get("/gallery/upload", (*km.AdminContext).GetGalleryUploads).
		Post("/gallery/upload", (*km.AdminContext).PostGalleryUpload).
		Delete("/gallery/upload", (*km.AdminContext).DeleteGalleryUpload).
		Get("/gallery/upload/url", (*km.AdminContext).GetGalleryUploadUrl).
		Get("/order", (*km.AdminContext).GetOrders).
		Put("/order", (*km.AdminContext).OverrideOrder).
		Put("/settings", (*km.AdminContext).UpdateGeneralSettings).
		Get("/", views.AdminView).
		/* Write new admin endpoints above. These two need to be the last admin endpoints. */
		Get("/:page", views.AdminView).
		Get("/:page/:subpage", views.AdminView)

	http.Handle("/", router)
}

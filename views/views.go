package views

import (
	"github.com/gocraft/web"
	"km"
	"html/template"
	"google.golang.org/appengine/log"
	"net/http"
	"settings"
	"entities"
	"strconv"
	"fmt"
)

type View struct {
	Title string
}

func HomeView(c *km.ServerContext, w web.ResponseWriter, r *web.Request) {
	var templates = template.Must(template.ParseGlob("views/template/*")) // cache this globally

	featuredCategories, err := entities.ListCategoriesByFeatured(c.Context, true)
	if err != nil {
		log.Errorf(c.Context, "Error getting featured categories: %+v", err)
		featuredCategories = []*entities.Category{}
	}

	log.Debugf(c.Context, "FeaturedCategories: %+v", featuredCategories)
	type HomeView struct {
		Title      string
		Categories []*entities.Category
	}

	p := HomeView{
		Title: settings.COMPANY_NAME + " | Home",
		Categories: featuredCategories,
	}

	err = templates.ExecuteTemplate(w, "home-page", p)
	if err != nil {
		log.Errorf(c.Context, "Error parsing home html file: %+v", err)
		c.ServeHTML(http.StatusInternalServerError, "Unexpected Error, please try again later.")
		return
	}
}

func ProductView(c *km.ServerContext, w web.ResponseWriter, r *web.Request) {
	var templates = template.Must(template.ParseGlob("views/template/*")) // cache this globally
	productIdStr := r.URL.Query().Get("p")
	if productIdStr == "" {
		productIdStr = r.PathParams["productId"]
	}

	productId, err := strconv.ParseInt(productIdStr, 10, 64)
	if err != nil {
		productId = int64(0)
	}

	log.Infof(c.Context, "Querying productId: %s", productId)
	productFound := true
	product, err := entities.GetProduct(c.Context, productId)
	if err != nil {
		log.Errorf(c.Context, "Error getting product: %+v", err)
		product = entities.NewProduct("Product not found")
		product.Description = "This product no longer exists."
		productFound = false
	}

	type ProductView struct {
		Title string
		Product *entities.Product
		ProductFound bool
		CanonicalUrl string
	}

	httpHeader := "http"
	if r.TLS != nil {
		httpHeader = "https"
	}
	p := ProductView {
		Title: settings.COMPANY_NAME + " | " + product.Name,
		Product: product,
		ProductFound: productFound,
		CanonicalUrl: fmt.Sprintf("%s://%s%s", httpHeader, r.Host, r.URL.Path),
	}

	log.Debugf(c.Context, "Canonical Url: %s", p.CanonicalUrl)

	if !productFound {
		w.WriteHeader(http.StatusNotFound)
	}

	err = templates.ExecuteTemplate(w, "product-page", p)
	if err != nil {
		log.Errorf(c.Context, "Error parsing store html file: %+v", err)
		c.ServeHTML(http.StatusInternalServerError, "Unexpected error, please try again later.")
		return
	}
}

func StoreView(c *km.ServerContext, w web.ResponseWriter, r *web.Request) {
	var templates = template.Must(template.ParseGlob("views/template/*")) // cache this globally
	category := r.URL.Query().Get("c")
	if category == "" {
		category = r.PathParams["category"]
	}

	log.Infof(c.Context, "Querying categories: %s", category)
	categories, err := entities.ListCategoriesByName(c.Context, category)
	if err != nil {
		log.Errorf(c.Context, "Error getting categories: %+v", err)
		c.ServeHTML(http.StatusInternalServerError, "Unexpected error, please try again later.")
		return
	}

	log.Infof(c.Context, "Categories found: %+v", categories)
	products, err := entities.GetProductsInCategories(c.Context, categories)
	if err != nil {
		log.Errorf(c.Context, "Error getting products: %+v", err)
		c.ServeHTML(http.StatusInternalServerError, "Unexpected error, please try again later.")
		return
	}

	log.Debugf(c.Context, "Products: %+v", products)
	for index, product := range products {
		log.Debugf(c.Context, "Product Thumbnail: %s", product.Thumbnail)
		if (index+1)%4 == 0 {
			product.Last = true
		}
	}

	featuredCategories, err := entities.ListCategoriesByFeatured(c.Context, true)
	if err != nil {
		log.Errorf(c.Context, "Error getting featured categories: %+v", err)
		featuredCategories = []*entities.Category{}
	}

	log.Debugf(c.Context, "FeaturedCategories: %+v", featuredCategories)

	type CategoryOption struct {
		Name string
		Selected bool
		Url string
	}

	options := make([]CategoryOption, len(featuredCategories))
	for index, cat := range featuredCategories {
		options[index] = CategoryOption{
			Name: cat.Name,
			Selected: category == cat.Name,
			Url: fmt.Sprintf("/store/%s", cat.Name),
		}
	}

	type ViewStore struct {
		Title string
		Products []*entities.Product
		Category string
		CategoryOptions []CategoryOption
		Categories []*entities.Category
	}

	p := ViewStore {
		Title: settings.COMPANY_NAME + " | Store",
		Products: products,
		Category: category,
		CategoryOptions: options,
	}

 	err = templates.ExecuteTemplate(w, "store-page", p)
	if err != nil {
		log.Errorf(c.Context, "Error parsing store html file: %+v", err)
		c.ServeHTML(http.StatusInternalServerError, "Unexpected error, please try again later.")
		return
	}
}

func AdminView(c *km.AdminContext, w web.ResponseWriter, r *web.Request) {
	p := View{
		Title: settings.COMPANY_NAME + " | Admin",
	}

	t, err := template.ParseFiles("views/admin.html") // cache this globally
	if err != nil {
		log.Errorf(c.Context, "Error parsing admin html file: %+v", err)
		c.ServeHTML(http.StatusInternalServerError, "Unexpected Error, please try again later.")
		return
	}

	t.Execute(w, p)
}

func RegisterView(c *km.ServerContext, w web.ResponseWriter, r *web.Request) {
	p := View{
		Title: settings.COMPANY_NAME + " | Register",
	}

	t, err := template.ParseFiles("views/register.html") // cache this globally
	if err != nil {
		log.Errorf(c.Context, "Error parsing register html file: %+v", err)
		c.ServeHTML(http.StatusInternalServerError, "Unexpected Error, please try again later.")
		return
	}

	t.Execute(w, p)
}

func LoginView(c *km.ServerContext, w web.ResponseWriter, r *web.Request) {
	p := View{
		Title: settings.COMPANY_NAME + " | Login",
	}

	t, err := template.ParseFiles("views/login.html") // cache this globally
	if err != nil {
		log.Errorf(c.Context, "Error parsing login html file: %+v", err)
		c.ServeHTML(http.StatusInternalServerError, "Unexpected Error, please try again later.")
		return
	}

	t.Execute(w, p)
}
func CartView(c *km.ServerContext, w web.ResponseWriter, r *web.Request) {
	var templates = template.Must(template.ParseGlob("views/template/*")) // cache this globally
	err := templates.ExecuteTemplate(w, "cart-page", nil)
	if err != nil {
		log.Errorf(c.Context, "Error parsing cart html file: %+v", err)
		c.ServeHTML(http.StatusInternalServerError, "Unexpected error, please try again later.")
		return
	}
}
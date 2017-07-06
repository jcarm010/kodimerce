package views

import (
	"github.com/gocraft/web"
	"github.com/jcarm010/kodimerce/km"
	"html/template"
	"google.golang.org/appengine/log"
	"net/http"
	"github.com/jcarm010/kodimerce/settings"
	"github.com/jcarm010/kodimerce/entities"
	"strconv"
	"fmt"
)

var templates = template.Must(template.New("").Funcs(fns).ParseGlob("views/templates/*")) //todo: cache this globally

type View struct {
	Title string
}

type OrderView struct {
	Title string
	Order *entities.Order `json:"order"`
}

var fns = template.FuncMap{
	"plus1": func(x int) int {
		return x + 1
	},
}

func HomeView(c *km.ServerContext, w web.ResponseWriter, r *web.Request) {

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

func ContactView(c *km.ServerContext, w web.ResponseWriter, r *web.Request) {
	p := struct {
		Title      string
	}{
		Title: settings.COMPANY_NAME + " | Contact",
	}

	err := templates.ExecuteTemplate(w, "contact-page", p)
	if err != nil {
		log.Errorf(c.Context, "Error parsing home html file: %+v", err)
		c.ServeHTML(http.StatusInternalServerError, "Unexpected Error, please try again later.")
		return
	}
}

func ReferralsView(c *km.ServerContext, w web.ResponseWriter, r *web.Request) {
	products, err := entities.ListProducts(c.Context)
	if err != nil {
		log.Errorf(c.Context, "Error listing products: %s", err)
		products = []*entities.Product{}
	}

	activeProducts := make([]*entities.Product, 0)
	for _, product := range products {
		if product.Active {
			activeProducts = append(activeProducts, product)
		}
	}

	p := struct {
		Title       string
		Products	[]*entities.Product
	}{
		Title: settings.COMPANY_NAME + " | Referrals",
		Products: activeProducts,
	}

	log.Infof(c.Context, "Products: %+v", products)
	err = templates.ExecuteTemplate(w, "referrals-page", p)
	if err != nil {
		log.Errorf(c.Context, "Error parsing home html file: %+v", err)
		c.ServeHTML(http.StatusInternalServerError, "Unexpected Error, please try again later.")
		return
	}
}

func ProductView(c *km.ServerContext, w web.ResponseWriter, r *web.Request) {
	productIdStr := r.URL.Query().Get("p")
	if productIdStr == "" {
		productIdStr = r.PathParams["productId"]
	}

	var selectedProduct *entities.Product
	productFound := true
	productId, err := strconv.ParseInt(productIdStr, 10, 64)
	if err != nil {
		log.Infof(c.Context, "Id is not a number, checking for product name.")
		selectedProduct, err = entities.GetProductByPath(c.Context, productIdStr)
	}else {
		log.Infof(c.Context, "Querying productId: %s", productId)
		selectedProduct, err = entities.GetProduct(c.Context, productId)
	}

	if err != nil {
		log.Errorf(c.Context, "Error getting product: %+v", err)
		selectedProduct = entities.NewProduct("Product not found")
		selectedProduct.Description = "This product no longer exists."
		productFound = false
	}

	type ProductView struct {
		Title string
		Product *entities.Product
		ProductFound bool
		CanonicalUrl string
		Domain string
	}

	httpHeader := "http"
	if r.TLS != nil {
		httpHeader = "https"
	}
	p := ProductView {
		Title: selectedProduct.Name,
		Product: selectedProduct,
		ProductFound: productFound,
		CanonicalUrl: fmt.Sprintf("%s://%s%s", httpHeader, r.Host, r.URL.Path),
		Domain: settings.ServerUrl(r.Request),
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

	type StoreView struct {
		Title string
		Products []*entities.Product
		Category string
		CategoryOptions []CategoryOption
		Categories []*entities.Category
		Domain string
	}

	title := settings.COMPANY_NAME + " | Store"
	if category != "" {
		title += " | " + category
	}

	p := StoreView{
		Title: title,
		Products: products,
		Category: category,
		CategoryOptions: options,
		Domain: settings.ServerUrl(r.Request),
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
	err := templates.ExecuteTemplate(w, "cart-page", struct{
		Title string
		TaxPercent float64
	}{
		Title: settings.COMPANY_NAME + " | Shopping Cart",
		TaxPercent: settings.TAX_PERCENT,
	})
	if err != nil {
		log.Errorf(c.Context, "Error parsing cart html file: %+v", err)
		c.ServeHTML(http.StatusInternalServerError, "Unexpected error, please try again later.")
		return
	}
}

func OrderReviewView(c *km.ServerContext, w web.ResponseWriter, r *web.Request) {
	orderIdStr := r.URL.Query().Get("id")
	if orderIdStr == "" {
		log.Errorf(c.Context, "Missing order id")
		c.ServeHTMLError(http.StatusBadRequest, "Could not find your order, please try again later.")
		return
	}

	orderId, err := strconv.ParseInt(orderIdStr, 10, 64)
	if err != nil {
		log.Errorf(c.Context, "Could not parse id: %+v", err)
		c.ServeHTMLError(http.StatusBadRequest, "Could not find your order, please try again later.")
		return
	}

	order, err := entities.GetOrder(c.Context, orderId)
	if err != nil {
		log.Errorf(c.Context, "Error finding order: %+v", err)
		c.ServeHTMLError(http.StatusBadRequest, "Could not find your order, please try again later.")
		return
	}

	log.Infof(c.Context, "Rendering orderId[%v] order[%+v]", orderId, order)

	err = templates.ExecuteTemplate(w, "order-review-page", struct{
		Title string
		Order *entities.Order
		TaxPercent float64
	}{
		Title: settings.COMPANY_NAME + " | Order Details",
		Order: order,
		TaxPercent: settings.TAX_PERCENT,
	})

	if err != nil {
		log.Errorf(c.Context, "Error parsing html file: %+v", err)
		c.ServeHTMLError(http.StatusInternalServerError, "Unexpected error, please try again later.")
		return
	}
}
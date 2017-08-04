package views

import (
	"github.com/gocraft/web"
	"github.com/jcarm010/kodimerce/km"
	"google.golang.org/appengine/log"
	"net/http"
	"github.com/jcarm010/kodimerce/settings"
	"github.com/jcarm010/kodimerce/entities"
	"strconv"
	"fmt"
	"html/template"
	"github.com/jcarm010/kodimerce/view"
	"github.com/jcarm010/feeds"
	"time"
	"google.golang.org/appengine/urlfetch"
	"io"
	"io/ioutil"
	"encoding/json"
)

var(
	CUSTOM_PAGES map[string]CustomPage
)

type CustomPage struct {
	TemplateName string `json:"template_name"`
	Title string `json:"title"`
	MetaDescription string `json:"meta_description"`
}

type OrderView struct {
	Title string
	Order *entities.Order `json:"order"`
}

func init() {
	customPages := struct{
		Pages map[string]CustomPage `json:"pages"`
	}{
		Pages: map[string]CustomPage{},
	}

	raw, err := ioutil.ReadFile("./custom-pages.json")
	if err == nil {
		err = json.Unmarshal(raw, &customPages)
		if err == nil {
			CUSTOM_PAGES = customPages.Pages
		}
	}
}

func HomeView(c *km.ServerContext, w web.ResponseWriter, r *web.Request) {
	featuredCategories, err := entities.ListCategoriesByFeatured(c.Context, true)
	if err != nil {
		log.Errorf(c.Context, "Error getting featured categories: %+v", err)
		featuredCategories = []*entities.Category{}
	}

	log.Debugf(c.Context, "FeaturedCategories: %+v", featuredCategories)
	title := settings.COMPANY_NAME
	if settings.META_TITLE_HOME != "" {
		title = settings.META_TITLE_HOME
	}

	p := struct{
		*view.View
		Categories []*entities.Category

	}{
		View: c.NewView(title, settings.META_DESCRIPTION_HOME),
		Categories: featuredCategories,
	}

	err = km.Templates.ExecuteTemplate(w, "home-page", p)
	if err != nil {
		log.Errorf(c.Context, "Error parsing home html file: %+v", err)
		c.ServeHTML(http.StatusInternalServerError, "Unexpected Error, please try again later.")
		return
	}
}

func ContactView(c *km.ServerContext, w web.ResponseWriter, r *web.Request) {
	p := struct {
		*view.View
	}{
		View: c.NewView("Contact | " + settings.COMPANY_NAME, settings.META_DESCRIPTION_CONTACT),
	}

	err := km.Templates.ExecuteTemplate(w, "contact-page", p)
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
		*view.View
		Products	[]*entities.Product
	}{
		View: c.NewView("Referrals | " + settings.COMPANY_NAME, settings.META_DESCRIPTION_REFERRALS),
		Products: activeProducts,
	}

	log.Infof(c.Context, "Products: %+v", products)
	err = km.Templates.ExecuteTemplate(w, "referrals-page", p)
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

	httpHeader := "http"
	if r.TLS != nil {
		httpHeader = "https"
	}
	p := struct {
		*view.View
		Product *entities.Product
		ProductFound bool
		CanonicalUrl string
		Domain string
	}{
		View: c.NewView(selectedProduct.Name + " | " + settings.COMPANY_NAME, selectedProduct.MetaDescription),
		Product: selectedProduct,
		ProductFound: productFound,
		CanonicalUrl: fmt.Sprintf("%s://%s%s", httpHeader, r.Host, r.URL.Path),
		Domain: settings.ServerUrl(r.Request),
	}

	log.Debugf(c.Context, "Canonical Url: %s", p.CanonicalUrl)

	if !productFound {
		w.WriteHeader(http.StatusNotFound)
	}

	err = km.Templates.ExecuteTemplate(w, "product-page", p)
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
	categories, err := entities.ListCategoriesByPath(c.Context, category)
	if err != nil {
		log.Errorf(c.Context, "Error getting categories by path: %+v", err)
		c.ServeHTML(http.StatusInternalServerError, "Unexpected error, please try again later.")
		return
	}

	if len(categories) == 0 && category != "" {
		categories, err = entities.ListCategoriesByName(c.Context, category)
		if err != nil {
			log.Errorf(c.Context, "Error getting categories by name: %+v", err)
			c.ServeHTML(http.StatusInternalServerError, "Unexpected error, please try again later.")
			return
		}
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
		Description string
	}

	options := make([]CategoryOption, len(featuredCategories))
	var metaDescription string = settings.META_DESCRIPTION_STORE
	var categoryName string = "Store"
	for index, cat := range featuredCategories {
		options[index] = CategoryOption{
			Name: cat.Name,
			Description: cat.Description,
			Selected: category == cat.Name || category == cat.Path,
			Url: fmt.Sprintf("/store/%s", cat.Path),
		}

		if options[index].Selected {
			metaDescription = cat.MetaDescription
			categoryName = cat.Name
		}
	}

	var title string = categoryName + " | " + settings.COMPANY_NAME
	p := struct {
		*view.View
		Products []*entities.Product
		Category string
		CategoryOptions []CategoryOption
		Categories []*entities.Category
		Domain string
	}{
		View: c.NewView(title, metaDescription),
		Products: products,
		Category: category,
		CategoryOptions: options,
		Domain: settings.ServerUrl(r.Request),
	}

 	err = km.Templates.ExecuteTemplate(w, "store-page", p)
	if err != nil {
		log.Errorf(c.Context, "Error parsing store html file: %+v", err)
		c.ServeHTML(http.StatusInternalServerError, "Unexpected error, please try again later.")
		return
	}
}

func AdminView(c *km.AdminContext, w web.ResponseWriter, r *web.Request) {
	p := c.NewView("Admin | " + settings.COMPANY_NAME, "")
	t, err := template.ParseFiles("views/admin.html") // cache this globally
	if err != nil {
		log.Errorf(c.Context, "Error parsing admin html file: %+v", err)
		c.ServeHTML(http.StatusInternalServerError, "Unexpected Error, please try again later.")
		return
	}

	t.Execute(w, p)
}

func RegisterView(c *km.ServerContext, w web.ResponseWriter, r *web.Request) {
	p := c.NewView("Register | " + settings.COMPANY_NAME, "")

	t, err := template.ParseFiles("views/register.html") // cache this globally
	if err != nil {
		log.Errorf(c.Context, "Error parsing register html file: %+v", err)
		c.ServeHTML(http.StatusInternalServerError, "Unexpected Error, please try again later.")
		return
	}

	t.Execute(w, p)
}

func LoginView(c *km.ServerContext, w web.ResponseWriter, r *web.Request) {
	p := c.NewView("Login | " + settings.COMPANY_NAME, "")

	t, err := template.ParseFiles("views/login.html") // cache this globally
	if err != nil {
		log.Errorf(c.Context, "Error parsing login html file: %+v", err)
		c.ServeHTML(http.StatusInternalServerError, "Unexpected Error, please try again later.")
		return
	}

	t.Execute(w, p)
}
func CartView(c *km.ServerContext, w web.ResponseWriter, r *web.Request) {
	err := km.Templates.ExecuteTemplate(w, "cart-page", struct{
		*view.View
		TaxPercent float64
	}{
		View: c.NewView("Shopping Cart | " + settings.COMPANY_NAME, settings.META_DESCRIPTION_CART),
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

	err = km.Templates.ExecuteTemplate(w, "order-review-page", struct{
		*view.View
		Title string
		Order *entities.Order
		TaxPercent float64
	}{
		View: c.NewView("Order Details | " + settings.COMPANY_NAME, ""),
		Order: order,
		TaxPercent: settings.TAX_PERCENT,
	})

	if err != nil {
		log.Errorf(c.Context, "Error parsing html file: %+v", err)
		c.ServeHTMLError(http.StatusInternalServerError, "Unexpected error, please try again later.")
		return
	}
}

func BlogView(c *km.ServerContext, w web.ResponseWriter, r *web.Request){
	posts, err := entities.ListPosts(c.Context, true, -1)
	if err != nil {
		log.Errorf(c.Context, "Error getting posts: %s", err)
		c.ServeHTML(http.StatusInternalServerError, "Unexpected error loading posts.")
		return
	}

	//sort.Sort(entities.ByNewestFirst(posts))
	err = km.Templates.ExecuteTemplate(w, "blog-page", struct {
		*view.View
		Posts []*entities.Post
	}{
		View: c.NewView("Blog | " + settings.COMPANY_NAME, settings.META_DESCRIPTION_BLOG),
		Posts: posts,
	})

	if err != nil {
		log.Errorf(c.Context, "Error parsing html file: %+v", err)
		c.ServeHTMLError(http.StatusInternalServerError, "Unexpected error, please try again later.")
		return
	}
}

func GetDynamicPage(c *km.ServerContext, w web.ResponseWriter, r *web.Request){
	postPath := r.PathParams["post"]
	log.Infof(c.Context, "Serving Post: %s", postPath)
	if customPage, exist := CUSTOM_PAGES[postPath] ; exist {
		log.Infof(c.Context, "Custom Page found: %+v", customPage)
		v := c.NewView(customPage.Title, customPage.MetaDescription)
		err := km.Templates.ExecuteTemplate(w, customPage.TemplateName, v)
		if err != nil {
			log.Errorf(c.Context, "Error parsing html file: %+v", err)
			c.ServeHTMLError(http.StatusInternalServerError, "Unexpected error, please try again later.")
		}

		return
	}

	post, err := entities.GetPostByPath(c.Context, postPath)
	if err != nil && err != entities.ErrPostNotFound {
		log.Errorf(c.Context, "Error getting post: %+v", err)
		c.ServeHTMLError(http.StatusInternalServerError, "Unexpected error, please try again later.")
		return
	}

	if err != entities.ErrPostNotFound {
		servePost(c, w, r, post)
		return
	}

	page, err := entities.GetPageByPath(c.Context, postPath)
	if err != nil && err != entities.ErrPageNotFound {
		log.Errorf(c.Context, "Error getting page: %+v", err)
		c.ServeHTMLError(http.StatusInternalServerError, "Unexpected error, please try again later.")
		return
	}

	if err != entities.ErrPageNotFound {
		servePage(c, w, r, page)
		return
	}

	c.ServeHTMLError(http.StatusNotFound, "The page you were looking for does not exist.")
}

func servePost(c *km.ServerContext, w web.ResponseWriter, r *web.Request, post *entities.Post) {
	posts, err := entities.ListPosts(c.Context, true, 10)
	if err != nil {
		log.Errorf(c.Context, "Error getting previous posts: %+v", err)
		posts = make([]*entities.Post, 0)
	}

	httpHeader := "http"
	if r.TLS != nil {
		httpHeader = "https"
	}
	err = km.Templates.ExecuteTemplate(w, "post-page", struct{
		*view.View
		CanonicalUrl string
		Post         *entities.Post
		LatestPosts  []*entities.Post
		AboutBlog    string
	}{
		View:         c.NewView(post.Title + " | " + settings.COMPANY_NAME, post.MetaDescription),
		CanonicalUrl: fmt.Sprintf("%s://%s%s", httpHeader, r.Host, r.URL.Path),
		Post:         post,
		LatestPosts:  posts,
		AboutBlog:    settings.DESCRIPTION_BLOG_ABOUT,
	})

	if err != nil {
		log.Errorf(c.Context, "Error parsing html file: %+v", err)
		c.ServeHTMLError(http.StatusInternalServerError, "Unexpected error, please try again later.")
	}
}

func servePage(c *km.ServerContext, w web.ResponseWriter, r *web.Request, page *entities.Page)  {
	if page.Provider == entities.PROVIDER_SHALLOW_MIRROR {
		resp, err := urlfetch.Client(c.Context).Get(page.ShallowMirrorUrl)
		if err != nil {
			log.Errorf(c.Context, "Error fetching mirrored page: %+v", err)
			c.ServeHTMLError(http.StatusInternalServerError, "Unexpected error, please try again later.")
			return
		}

		defer resp.Body.Close()
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			log.Errorf(c.Context, "Error copying mirrored page response: %+v", err)
			c.ServeHTMLError(http.StatusInternalServerError, "Unexpected error, please try again later.")
			return
		}
	} else {
		log.Errorf(c.Context, "Page provider is not supported: %+v", page)
		c.ServeHTMLError(http.StatusInternalServerError, "Unexpected error, please try again later.")
	}
}

func GetBlogRss(c *km.ServerContext, w web.ResponseWriter, r *web.Request){
	posts, err := entities.ListPosts(c.Context, true, -1)
	if err != nil {
		log.Errorf(c.Context, "Error getting posts: %s", err)
		c.ServeHTML(http.StatusInternalServerError, "Unexpected error loading posts.")
		return
	}

	serverUrl := settings.ServerUrl(r.Request)
	//sort.Sort(entities.ByNewestFirst(posts))
	now := time.Now()
	feed := &feeds.Feed{
		Title:       settings.COMPANY_NAME + " Blog",
		Link:        &feeds.Link{Href: serverUrl + "/blog"},
		Description: settings.META_DESCRIPTION_BLOG,
		Author:      &feeds.Author{Name: settings.COMPANY_NAME, Email: settings.COMPANY_SUPPORT_EMAIL},
		Created:     now,
	}
	feed.Items = make([]*feeds.Item, len(posts))
	for index, post := range posts {
		postUrl := serverUrl + "/" + post.Path
		feed.Items[index] = &feeds.Item{
			Id: 		 postUrl,
			Title:       post.Title,
			Link:        &feeds.Link{Href: postUrl},
			Description: post.MetaDescription,
			Created:     post.PublishedDate,
		}
	}

	rss, err := feed.ToRss()
	if err != nil {
		log.Errorf(c.Context, "Error loading rss feed: %s", err)
		c.ServeHTML(http.StatusInternalServerError, "Unexpected error loading rss feed.")
		return
	}

	w.Header().Add("Content-Type:","text/xml; charset=utf-8")
	w.Write([]byte(rss))
}

func ThankYouView(c *km.ServerContext, w web.ResponseWriter, r *web.Request) {
	orderIdStr := r.URL.Query().Get("order")
	var order *entities.Order
	if orderIdStr != "" {
		orderId, err := strconv.ParseInt(orderIdStr, 10, 64)
		if err != nil {
			log.Errorf(c.Context, "Could not parse id: %+v", err)
			c.ServeHTMLError(http.StatusBadRequest, "Could not find your order, please try again later.")
			return
		}

		order, err = entities.GetOrder(c.Context, orderId)
		if err != nil {
			log.Errorf(c.Context, "Error finding order: %+v", err)
			c.ServeHTMLError(http.StatusBadRequest, "Could not find your order, please try again later.")
			return
		}

		log.Infof(c.Context, "Rendering orderId[%v] order[%+v]", orderId, order)
	}

	err := km.Templates.ExecuteTemplate(w, "thank-you-page", struct{
		*view.View
		Order *entities.Order `json:"order"`
		HasOrder bool `json:"has_order"`
	}{
		View: c.NewView("Thank You | " + settings.COMPANY_NAME, "Thank you."),
		Order: order,
		HasOrder: order != nil,
	})

	if err != nil {
		log.Errorf(c.Context, "Error parsing html file: %+v", err)
		c.ServeHTMLError(http.StatusInternalServerError, "Unexpected error, please try again later.")
		return
	}
}

func GalleriesView(c *km.ServerContext, w web.ResponseWriter, r *web.Request) {
	galleries, err := entities.ListGalleries(c.Context, true, -1)
	if err != nil {
		log.Errorf(c.Context, "Error getting galleries: %+v", err)
		c.ServeHTMLError(http.StatusInternalServerError, "Unexpected error, please try again later.")
		return
	}

	err = km.Templates.ExecuteTemplate(w, "galleries-page", struct{
		*view.View
		Galleries []*entities.Gallery `json:"galleries"`
	}{
		View: c.NewView("Galleries | " + settings.COMPANY_NAME, settings.META_DESCRIPTION_GALLERIES),
		Galleries: galleries,
	})

	if err != nil {
		log.Errorf(c.Context, "Error parsing html file: %+v", err)
		c.ServeHTMLError(http.StatusInternalServerError, "Unexpected error, please try again later.")
		return
	}
}

func GalleryView(c *km.ServerContext, w web.ResponseWriter, r *web.Request) {
	galleryPath := r.PathParams["galleryPath"]
	gallery, err := entities.GetGalleryByPath(c.Context, galleryPath)
	if err == entities.ErrGalleryNotFound {
		c.ServeHTMLError(http.StatusNotFound, "The gallery you were looking for does not exist.")
		return
	}

	if err != nil {
		log.Errorf(c.Context, "Error getting gallery: %+v", err)
		c.ServeHTMLError(http.StatusInternalServerError, "Unexpected error, please try again later.")
		return
	}

	err = km.Templates.ExecuteTemplate(w, "gallery-page", struct{
		*view.View
		Gallery *entities.Gallery `json:"gallery"`
	}{
		View: c.NewView(gallery.Title + " | " + settings.COMPANY_NAME, gallery.MetaDescription),
		Gallery: gallery,
	})

	if err != nil {
		log.Errorf(c.Context, "Error parsing html file: %+v", err)
		c.ServeHTMLError(http.StatusInternalServerError, "Unexpected error, please try again later.")
		return
	}
}
package km

import (
	"github.com/gocraft/web"
	"github.com/jcarm010/kodimerce/entities"
	"google.golang.org/appengine/log"
	"net/http"
	"strconv"
	"strings"
	"google.golang.org/appengine/blobstore"
	"google.golang.org/appengine"
	"fmt"
	"sort"
	"time"
)

type AdminContext struct {
	*ServerContext
	User *entities.User
}

func (c *AdminContext) Auth(w web.ResponseWriter, r *web.Request, next web.NextMiddlewareFunc) {
	cookies := r.Cookies()
	sessionToken := ""
	for _, cookie := range cookies {
		if cookie.Name == "km-session" {
			sessionToken = cookie.Value
		}
	}

	if sessionToken == "" {
		if r.Method == "GET" {
			http.Redirect(w, r.Request, "/login", http.StatusTemporaryRedirect)
		}else {
			c.ServeJson(http.StatusUnauthorized, "Missing session.")
		}
		return
	}


	userSession, err := entities.GetUserSession(c.Context, sessionToken)
	if err != nil {
		log.Errorf(c.Context, "Error getting session: %+v", err)
		if r.Method == "GET" {
			http.Redirect(w, r.Request, "/login", http.StatusTemporaryRedirect)
		}else {
			c.ServeJson(http.StatusUnauthorized, "Session not found.")
		}

		return
	}

	user, err := entities.GetUser(c.Context, userSession.Email)
	if err != nil {
		log.Errorf(c.Context, "Error getting user: %+v", err)
		if r.Method == "GET" {
			http.Redirect(w, r.Request, "/login", http.StatusTemporaryRedirect)
		}else {
			c.ServeJson(http.StatusUnauthorized, "Session not found.")
		}
		return
	}

	if user.UserType != "admin" {
		log.Errorf(c.Context, "User is not admin: %+v", user)
		if r.Method == "GET" {
			http.Redirect(w, r.Request, "/login", http.StatusTemporaryRedirect)
		}else {
			c.ServeJson(http.StatusForbidden, "Not allowed.")
		}
		return
	}

	log.Debugf(c.Context, "Authenticated user: %+v", user.Email)
	c.User = user
	next(w, r)
}

func (c *AdminContext) GetProducts(w web.ResponseWriter, r *web.Request) {
	products, err := entities.ListProducts(c.Context)
	if err != nil {
		log.Errorf(c.Context, "Error getting products: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Unexpected error getting products.")
		return
	}

	c.ServeJson(http.StatusOK, products)
}

func (c *AdminContext) CreateProduct(w web.ResponseWriter, r *web.Request) {
	name := r.URL.Query().Get("name")
	log.Infof(c.Context, "Creating product: %+v", name)
	if name == "" {
		c.ServeJson(http.StatusBadRequest, "Name cannot be empty")
		return
	}

	product, err := entities.CreateProduct(c.Context, name)
	if err != nil {
		log.Errorf(c.Context, "Error creating product: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Unexpected error creating product.")
		return
	}

	c.ServeJson(http.StatusOK, product)
}

func (c *AdminContext) UpdateProduct(w web.ResponseWriter, r *web.Request) {
	product := &entities.Product{}
	err := c.ParseJsonRequest(product)
	if err != nil {
		log.Errorf(c.Context, "Failed to parse parse product: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Failed to parse product.")
		return
	}

	log.Infof(c.Context, "Updating product: %+v", product)
	if product.Id == 0 {
		c.ServeJson(http.StatusBadRequest, "Id cannot be empty")
		return
	}
	if product.Name == "" {
		c.ServeJson(http.StatusBadRequest, "Name cannot be empty")
		return
	}
	log.Infof(c.Context, "PriceCents: %+v", product.PriceCents)
	log.Infof(c.Context, "Active: %+v", product.Active)
	log.Infof(c.Context, "isInfinite: %+v", product.IsInfinite)
	log.Infof(c.Context, "noShipping: %+v", product.NoShipping)
	log.Infof(c.Context, "needsDate: %+v", product.NeedsDate)
	log.Infof(c.Context, "needsTime: %+v", product.NeedsTime)
	log.Infof(c.Context, "needsPickupLocation: %+v", product.NeedsPickupLocation)
	log.Infof(c.Context, "hasPricingOptions: %+v", product.HasPricingOptions)
	if product.AvailableTimes == nil {
		product.AvailableTimes = make([]entities.AvailableTime, 0)
	}

	sort.Sort(entities.ByAvailableTime(product.AvailableTimes))
	log.Infof(c.Context, "availableTimes: %+v:", product.AvailableTimes)
	if product.PricingOptions == nil {
		product.PricingOptions = make([]entities.PricingOption, 0)
	}

	sort.Sort(entities.ByCheapestPrice(product.PricingOptions))
	log.Infof(c.Context, "pricingOptions: %+v:", product.PricingOptions)
	if product.Path == "" {
		product.Path = fmt.Sprintf("%s", product.Id)
	}

	if product.Pictures == nil {
		product.Pictures = []string{}
	}

	err = entities.UpdateProduct(c.Context, product)
	if err != nil {
		log.Errorf(c.Context, "Error storing product: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Unexpected value storing product")
		return
	}
}

func (c *AdminContext) GetCategory(w web.ResponseWriter, r *web.Request) {
	categories, err := entities.ListCategories(c.Context)
	if err != nil {
		log.Errorf(c.Context, "Error getting categories: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Unexpected error getting categories.")
		return
	}

	c.ServeJson(http.StatusOK, categories)
}

func (c *AdminContext) GetCategoryProduct(w web.ResponseWriter, r *web.Request) {
	categoryProducts, err := entities.GetCategoryProducts(c.Context)
	if err != nil {
		log.Errorf(c.Context, "Error getting category products: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Unexpected error getting categories.")
		return
	}

	c.ServeJson(http.StatusOK, categoryProducts)
}

func (c *AdminContext) CreateCategory(w web.ResponseWriter, r *web.Request) {
	name := r.URL.Query().Get("name")
	log.Infof(c.Context, "Creating category: %+v", name)
	if name == "" {
		c.ServeJson(http.StatusBadRequest, "Name cannot be empty")
		return
	}

	category, err := entities.CreateCategory(c.Context, name)
	if err != nil {
		log.Errorf(c.Context, "Error creating category: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Unexpected error creating category.")
		return
	}

	c.ServeJson(http.StatusOK, category)
}

func (c *AdminContext) UpdateCategory(w web.ResponseWriter, r *web.Request) {
	category := &entities.Category{}
	err := c.ParseJsonRequest(category)
	if err != nil {
		log.Errorf(c.Context, "Failed to parse parse category: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Failed to parse category.")
		return
	}

	defer r.Body.Close()
	log.Infof(c.Context, "Updating category: %+v", category)
	err = entities.UpdateCategory(c.Context, category)
	if err != nil {
		log.Errorf(c.Context, "Error storing category: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Unexpected value storing category")
		return
	}
}

func (c *AdminContext) SetCategoryProducts(w web.ResponseWriter, r *web.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Errorf(c.Context, "Failed to parse category products: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Failed to parse parameters.")
		return
	}

	toAdd := r.FormValue("to_add")
	colSeparated := strings.Split(toAdd, ",")
	categoryProducts := make([]*entities.CategoryProduct, 0)
	for _, str := range colSeparated {
		log.Debugf(c.Context, "%s", str)
		tokens := strings.Split(str, ":")
		if len(tokens) < 2 {
			log.Errorf(c.Context, "Could not separate tokens")
			continue;
		}

		categoryId, err := strconv.ParseInt(tokens[0], 10, 64)
		if err != nil {
			log.Errorf(c.Context, "Error parsing categoryId: %+v", err)
			continue
		}

		productId, err := strconv.ParseInt(tokens[1], 10, 64)
		if err != nil {
			log.Errorf(c.Context, "Error parsing productId: %+v", err)
			continue
		}

		log.Debugf(c.Context, "%v => %v", categoryId, productId)
		categoryProducts = append(categoryProducts, &entities.CategoryProduct{
			ProductId: productId,
			CategoryId: categoryId,
		})
	}

	err = entities.SetCategoryProducts(c.Context, categoryProducts)
	if err != nil {
		log.Errorf(c.Context, "Error storing category products: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Unexpected value storing category products")
		return
	}
}

func (c *AdminContext) UnsetCategoryProducts(w web.ResponseWriter, r *web.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Errorf(c.Context, "Failed to parse category products: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Failed to parse parameters.")
		return
	}

	categoryProducts := make([]*entities.CategoryProduct, 0)
	for categoryIdStr, productIdsStr := range r.Form {
		log.Debugf(c.Context, "%s => %s", categoryIdStr, productIdsStr)
		categoryId, err := strconv.ParseInt(categoryIdStr, 10, 64)
		if err != nil {
			log.Errorf(c.Context, "Error parsing categoryId: %+v", err)
			continue
		}

		for _, productIdStr := range productIdsStr {
			productId, err := strconv.ParseInt(productIdStr, 10, 64)
			if err != nil {
				log.Errorf(c.Context, "Error parsing productId: %+v", err)
				continue
			}

			categoryProducts = append(categoryProducts, &entities.CategoryProduct{
				CategoryId: categoryId,
				ProductId: productId,
			})
		}
	}

	err = entities.UnsetCategoryProducts(c.Context, categoryProducts)
	if err != nil {
		log.Errorf(c.Context, "Error deleting category products: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Unexpected value deleting category products")
		return
	}
}

func (c *AdminContext) GetGalleryUploadUrl(w web.ResponseWriter, r *web.Request) {
	uploadURL, err := blobstore.UploadURL(c.Context, "/admin/gallery/upload", nil)
	if err != nil {
		log.Errorf(c.Context, "Error creating blobstore url")
		c.ServeJson(http.StatusInternalServerError, "Unexpected error creating upload url")
		return
	}

	log.Infof(c.Context, "Upload url: %+v", uploadURL)
	c.ServeJson(http.StatusOK, uploadURL.String())
}

func (c *AdminContext) PostGalleryUpload(w web.ResponseWriter, r *web.Request) {
	blobs, _, err := blobstore.ParseUpload(r.Request)
	if err != nil {
		log.Errorf(c.Context, "Error parsing upload: %+v", err)
		c.ServeJson(http.StatusBadRequest, "Could not parse upload.")
		return
	}
	file := blobs["file"]
	if len(file) == 0 {
		log.Errorf(c.Context, "No file uploaded")
		c.ServeJson(http.StatusBadRequest, "No file uploaded")
		return
	}
	//key := file[0].BlobKey
	c.ServeJson(http.StatusOK, file)
}

func (c *AdminContext) DeleteGalleryUpload(w web.ResponseWriter, r *web.Request) {
	key := r.URL.Query().Get("k")
	if key == "" {
		c.ServeHTML(http.StatusNotFound, "Upload not found")
		return
	}

	err := blobstore.Delete(c.Context, appengine.BlobKey(key))
	if err != nil {
		log.Errorf(c.Context, "Error removing file: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Unexpected error removing file")
		return
	}
}

func (c *AdminContext) GetGalleryUploads(w web.ResponseWriter, r *web.Request) {
	blobs, err := entities.ListUploads(c.Context)
	if err != nil {
		log.Errorf(c.Context, "Error fetching blobs: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Unexpected error getting uploads")
		return
	}

	c.ServeJson(http.StatusOK, blobs)
}

func (c *ServerContext) GetGalleryUploadByName(w web.ResponseWriter, r *web.Request) {
	name := r.PathParams["name"]
	if name == "" {
		log.Errorf(c.Context, "Name not provided.")
		c.ServeHTML(http.StatusNotFound, "Upload not found")
		return
	}

	upload, err := entities.GetUploadByName(c.Context, name)
	if err != nil {
		log.Errorf(c.Context, "Error getting upload: %s", err)
		c.ServeJson(http.StatusNotFound, "Upload not found.")
		return
	}

	contentType := upload.ContentType
	//Content-Type: text/html
	w.Header().Add("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", name))
	w.Header().Add("Cache-Control", "max-age=2593000")
	w.Header().Add("Content-Type", contentType)
	blobstore.Send(w, upload.BlobKey)
}

func (c *ServerContext) GetGalleryUpload(w web.ResponseWriter, r *web.Request) {
	key := r.PathParams["key"]
	if key == "" {
		key = r.URL.Query().Get("k")
		if key == "" {
			c.ServeHTML(http.StatusNotFound, "Upload not found")
			return
		}
	}

	upload, err := entities.GetUpload(c.Context, appengine.BlobKey(key))
	if err != nil {
		log.Errorf(c.Context, "Error getting upload: %s", err)
		c.ServeJson(http.StatusNotFound, "Upload not found.")
		return
	}

	name := upload.Filename
	contentType := upload.ContentType
	cacheUntil := time.Now().AddDate(0, 2, 0).Format(http.TimeFormat)
	//Content-Type: text/html
	w.Header().Add("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", name))
	w.Header().Add("Cache-Control", "max-age=2593000")
	w.Header().Add("Content-Type", contentType)
	w.Header().Set("Expires", cacheUntil)
	blobstore.Send(w, upload.BlobKey)
}

func (c *AdminContext) GetOrders(w web.ResponseWriter, r *web.Request) {
	orders, err := entities.ListOrders(c.Context)
	if err != nil {
		log.Errorf(c.Context, "Error fetching orders: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Unexpected error getting orders")
		return
	}

	c.ServeJson(http.StatusOK, orders)
}

func (c *AdminContext) OverrideOrder(w web.ResponseWriter, r *web.Request){
	err := r.ParseForm()
	if err != nil {
		log.Errorf(c.Context, "Error parsing form: %+v", err)
		c.ServeJson(http.StatusBadRequest, "Could not understand the request. Please try again later.")
		return
	}

	idStr := r.FormValue("id")
	shippingName := r.FormValue("shipping_name")
	shippingLine1 := r.FormValue("shipping_line_1")
	shippingLine2 := r.FormValue("shipping_line_2")
	city := r.FormValue("city")
	state := r.FormValue("state")
	postalCode := r.FormValue("postal_code")
	countryCode := r.FormValue("country_code")
	email := r.FormValue("email")
	phone := r.FormValue("phone")
	checkoutStep := r.FormValue("checkout_step")
	paypalPayerId := r.FormValue("paypal_payer_id")
	addressVerifiedStr := r.FormValue("address_verified")
	status := r.FormValue("status")

	log.Infof(c.Context, "Updating order idStr[%s] shippingName[%s] shippingLine1[%s] shippingLine2[%s] city[%s] state[%s] postalCode[%s] countryCode[%s] email[%s] phone[%s] checkoutStep[%s] paypalPayerId[%s] addressVerifiedStr[%s] status[%s]",
		idStr, shippingName, shippingLine1, shippingLine2, city, state, postalCode, countryCode, email, phone, checkoutStep, paypalPayerId, addressVerifiedStr, status)

	if shippingName == "" {
		log.Errorf(c.Context, "Missing shipping name")
		c.ServeJson(http.StatusBadRequest, "Missing name")
		return
	}

	if email == "" {
		log.Errorf(c.Context, "Missing email")
		c.ServeJson(http.StatusBadRequest, "Missing email")
		return
	}

	orderId, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Errorf(c.Context, "Error parsing id: %+v", err)
		c.ServeJson(http.StatusBadRequest, "Could not understand the request. Please try again later.")
		return
	}

	order, err := entities.GetOrder(c.Context, orderId)
	if err != nil {
		log.Errorf(c.Context, "Error finding order: %+v", err)
		c.ServeJson(http.StatusBadRequest, "Could not find order. Please try again later.")
		return
	}

	order.ShippingName = shippingName
	order.ShippingLine1 = shippingLine1
	order.ShippingLine2 = shippingLine2
	order.City = city
	order.State = state
	order.PostalCode = postalCode
	order.CountryCode = countryCode
	order.Email = email
	order.Phone = phone
	order.CheckoutStep = checkoutStep
	order.PaypalPayerId = paypalPayerId
	order.AddressVerified = addressVerifiedStr == "true"
	order.Status = status

	err = entities.UpdateOrder(c.Context, order)
	if err != nil {
		log.Errorf(c.Context, "Error updating order: %+v", err)
		c.ServeJson(http.StatusBadRequest, "Could not update order. Please try again later.")
		return
	}

	c.ServeJson(http.StatusOK, "")
}

func (c *AdminContext) GetPosts(w web.ResponseWriter, r *web.Request){
	posts, err := entities.ListPosts(c.Context, false, -1)
	if err != nil {
		log.Errorf(c.Context, "Error loading posts: %s", err)
		c.ServeJson(http.StatusInternalServerError, "Unexpected error loading posts.")
		return
	}

	c.ServeJson(http.StatusOK, posts)
}

func (c *AdminContext) CreatePost(w web.ResponseWriter, r *web.Request){
	data := struct {
		Title string `json:"title"`
	}{}

	err := c.ParseJsonRequest(&data)
	if err != nil {
		log.Errorf(c.Context, "Could not parse request: %+v", err)
		c.ServeJson(http.StatusBadRequest, "Could not parse request.")
		return
	}

	log.Infof(c.Context, "Creating post with title: %s", data.Title)
	if data.Title == "" {
		c.ServeJson(http.StatusBadRequest, "A post needs a title.")
		return
	}

	post, err := entities.CreatePost(c.Context, data.Title)
	if err != nil {
		log.Errorf(c.Context, "Error creating post: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Unexpected error creating post.")
		return
	}

	c.ServeJson(http.StatusOK, post)
}

func (c *AdminContext) UpdatePost(w web.ResponseWriter, r *web.Request){
	data := struct {
		Post *entities.Post `json:"post"`
	}{}

	err := c.ParseJsonRequest(&data)
	if err != nil {
		log.Errorf(c.Context, "Could not parse request: %+v", err)
		c.ServeJson(http.StatusBadRequest, "Could not parse post.")
		return
	}

	log.Infof(c.Context, "Updating post: %+v", data.Post)
	err = entities.UpdatePost(c.Context, data.Post)
	if err != nil {
		log.Errorf(c.Context, "Error updating post: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Unexpected error updating post.")
		return
	}

	c.ServeJson(http.StatusOK, "")
}

func (c *AdminContext) GetGalleries(w web.ResponseWriter, r *web.Request){
	galleries, err := entities.ListGalleries(c.Context, false, -1)
	if err != nil {
		log.Errorf(c.Context, "Error listing galleries: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Unexpected error getting galleries.")
		return
	}

	c.ServeJson(http.StatusOK, galleries)
}


func (c *AdminContext) CreateGallery(w web.ResponseWriter, r *web.Request){
	data := struct {
		Title string `json:"title"`
	}{}

	err := c.ParseJsonRequest(&data)
	if err != nil {
		log.Errorf(c.Context, "Could not parse request: %+v", err)
		c.ServeJson(http.StatusBadRequest, "Could not parse request.")
		return
	}

	gallery, err := entities.CreateGallery(c.Context, data.Title)
	if err != nil {
		log.Errorf(c.Context, "Error creating gallery: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Unexpected error creating gallery.")
		return
	}

	c.ServeJson(http.StatusOK, gallery)
}

func (c *AdminContext) UpdateGallery(w web.ResponseWriter, r *web.Request){
	data := struct {
		Gallery *entities.Gallery `json:"gallery"`
	}{}

	err := c.ParseJsonRequest(&data)
	if err != nil {
		log.Errorf(c.Context, "Could not parse request: %+v", err)
		c.ServeJson(http.StatusBadRequest, "Could not parse request.")
		return
	}

	err = entities.UpdateGallery(c.Context, data.Gallery)
	if err != nil {
		log.Errorf(c.Context, "Error updating gallery: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Unexpected error updating gallery.")
		return
	}

	c.ServeJson(http.StatusOK, "")
}

func (c *AdminContext) CreatePage(w web.ResponseWriter, r *web.Request){
	data := struct {
		Title string `json:"title"`
	}{}

	err := c.ParseJsonRequest(&data)
	if err != nil {
		log.Errorf(c.Context, "Could not parse request: %+v", err)
		c.ServeJson(http.StatusBadRequest, "Could not parse request.")
		return
	}

	page, err := entities.CreatePage(c.Context, data.Title)
	if err != nil {
		log.Errorf(c.Context, "Error creating page: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Unexpected error creating page.")
		return
	}

	c.ServeJson(http.StatusOK, page)
}

func (c *AdminContext) GetPages(w web.ResponseWriter, r *web.Request){
	pages, err := entities.ListPages(c.Context, false, -1)
	if err != nil {
		log.Errorf(c.Context, "Error listing pages: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Unexpected error getting pages.")
		return
	}

	c.ServeJson(http.StatusOK, pages)
}

func (c *AdminContext) UpdatePage(w web.ResponseWriter, r *web.Request){
	data := struct {
		Page *entities.Page `json:"page"`
	}{}

	err := c.ParseJsonRequest(&data)
	if err != nil {
		log.Errorf(c.Context, "Could not parse request: %+v", err)
		c.ServeJson(http.StatusBadRequest, "Could not parse request.")
		return
	}

	log.Infof(c.Context, "Updating page: %+v", data.Page)
	err = entities.UpdatePage(c.Context, data.Page)
	if err != nil {
		log.Errorf(c.Context, "Error updating page: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Unexpected error updating page.")
		return
	}

	c.ServeJson(http.StatusOK, "")
}
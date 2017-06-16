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
	"encoding/json"
	"sort"
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
	err := r.ParseForm()
	if err != nil {
		log.Errorf(c.Context, "Failed to parse update product: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Failed to parse product.")
		return
	}

	idStr := r.FormValue("id")
	name := r.FormValue("name")
	priceCentsStr := r.FormValue("price_cents")
	quantityStr := r.FormValue("quantity")
	activeStr := r.FormValue("active")
	picturesStr := r.FormValue("pictures")
	description := r.FormValue("description")
	isInfiniteStr := r.FormValue("is_infinite")
	needsDateStr := r.FormValue("needs_date")
	needsTimeStr := r.FormValue("needs_time")
	noShippingStr := r.FormValue("no_shipping")
	availableTimesStr := r.FormValue("available_times")
	log.Infof(c.Context, "Modifying product [%s] with values: name[%s] price_cents[%s] quantity[%s] active[%s] pictures[%s] description[%s]", idStr, name, priceCentsStr, quantityStr, activeStr, picturesStr, description)
	var id int64
	if idStr == "" {
		c.ServeJson(http.StatusBadRequest, "Id cannot be empty")
		return
	}else {
		id, err = strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.Errorf(c.Context, "Error parsing idStr: %+v", err)
			c.ServeJson(http.StatusBadRequest, "Invalid value for id")
			return
		}
	}

	log.Infof(c.Context, "Id: %+v", id)
	if name == "" {
		c.ServeJson(http.StatusBadRequest, "Name cannot be empty")
		return
	}

	var priceCents int64 = 0
	if priceCentsStr != "" {
		priceCents, err = strconv.ParseInt(priceCentsStr, 10, 64)
		if err != nil {
			log.Errorf(c.Context, "Error parsing priceCentsStr: %+v", err)
			c.ServeJson(http.StatusBadRequest, "Invalid value for price")
			return
		}
	}

	log.Infof(c.Context, "PriceCents: %+v", priceCents)
	var active bool = false
	if activeStr != "" {
		active, err = strconv.ParseBool(activeStr)
		if err != nil {
			log.Errorf(c.Context, "Error parsing activeStr: %+v", err)
			c.ServeJson(http.StatusBadRequest, "Invalid value for active")
			return
		}
	}

	log.Infof(c.Context, "Active: %+v", active)
	var isInfinite bool = false
	if isInfiniteStr != "" {
		isInfinite, err = strconv.ParseBool(isInfiniteStr)
		if err != nil {
			log.Errorf(c.Context, "Error parsing isInfiniteStr: %+v", err)
			c.ServeJson(http.StatusBadRequest, "Invalid value for is_infinite")
			return
		}
	}

	log.Infof(c.Context, "isInfinite: %+v", isInfinite)
	var noShipping bool = false
	if noShippingStr != "" {
		noShipping, err = strconv.ParseBool(noShippingStr)
		if err != nil {
			log.Errorf(c.Context, "Error parsing noShippingStr: %+v", err)
			c.ServeJson(http.StatusBadRequest, "Invalid value for no_shipping")
			return
		}
	}

	log.Infof(c.Context, "noShipping: %+v", noShipping)
	var needsDate bool = false
	if needsDateStr != "" {
		needsDate, err = strconv.ParseBool(needsDateStr)
		if err != nil {
			log.Errorf(c.Context, "Error parsing needsDateStr: %+v", err)
			c.ServeJson(http.StatusBadRequest, "Invalid value for needs_date")
			return
		}
	}

	log.Infof(c.Context, "needsDate: %+v", needsDate)
	var needsTime bool = false
	if needsTimeStr != "" {
		needsTime, err = strconv.ParseBool(needsTimeStr)
		if err != nil {
			log.Errorf(c.Context, "Error parsing needsTimeStr: %+v", err)
			c.ServeJson(http.StatusBadRequest, "Invalid value for needs_time")
			return
		}
	}

	log.Infof(c.Context, "needsDate: %+v", needsDate)
	availableTimes := make([]entities.AvailableTime, 0)
	if availableTimesStr != "" {
		err = json.Unmarshal([]byte(availableTimesStr), &availableTimes)
		if err != nil {
			log.Errorf(c.Context, "Error parsing availableTimesStr: %+v", err)
			c.ServeJson(http.StatusBadRequest, "Invalid value for available_times")
			return
		}
	}

	sort.Sort(entities.ByAvailableTime(availableTimes))
	log.Infof(c.Context, "availableTimesStr: %+v:", availableTimesStr)
	log.Infof(c.Context, "availableTimes: %+v:", availableTimes)
	var quantity int = 0
	if quantityStr != "" {
		quantity64, err := strconv.ParseInt(quantityStr, 10, 32)
		if err != nil {
			log.Errorf(c.Context, "Error parsing quantityStr: %+v", err)
			c.ServeJson(http.StatusBadRequest, "Invalid value for quantity")
			return
		}
		quantity = int(quantity64)
	}

	log.Infof(c.Context, "Quantity: %+v", quantity)
	product := entities.NewProduct(name)
	product.Id = id
	product.Active = active
	product.PriceCents = priceCents
	product.Quantity = quantity
	product.Description = description
	product.IsInfinite = isInfinite
	product.NoShipping = noShipping
	product.NeedsDate = needsDate
	product.NeedsTime = needsTime
	product.AvailableTimes = availableTimes
	if picturesStr != "" {
		product.Pictures = strings.Split(picturesStr,",")
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
	err := r.ParseForm()
	if err != nil {
		log.Errorf(c.Context, "Failed to parse update category: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Failed to parse category.")
		return
	}

	idStr := r.FormValue("id")
	name := r.FormValue("name")
	description := r.FormValue("description")
	featuredStr := r.FormValue("featured")
	thumbnail := r.FormValue("thumbnail")
	log.Infof(c.Context, "Modifying category [%s] with values: name[%s] description[%s] featuredStr[%s] thumbnail[%s]", idStr, name, description, featuredStr, thumbnail)
	var id int64
	if idStr == "" {
		c.ServeJson(http.StatusBadRequest, "Id cannot be empty")
		return
	}else {
		id, err = strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.Errorf(c.Context, "Error parsing idStr: %+v", err)
			c.ServeJson(http.StatusBadRequest, "Invalid value for id")
			return
		}
	}

	log.Infof(c.Context, "Id: %+v", id)
	if name == "" {
		c.ServeJson(http.StatusBadRequest, "Name cannot be empty")
		return
	}

	featured, err := strconv.ParseBool(featuredStr)
	if err != nil {
		log.Errorf(c.Context, "Error parsing featured parameter: %+v", err)
		featured = false
	}

	category := entities.NewCategory(name)
	category.Id = id
	category.Description = description
	category.Thumbnail = thumbnail
	category.Featured = featured
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


func (c *ServerContext) GetGalleryUpload(w web.ResponseWriter, r *web.Request) {
	key := r.URL.Query().Get("k")
	if key == "" {
		c.ServeHTML(http.StatusNotFound, "Upload not found")
		return
	}

	blobstore.Send(w, appengine.BlobKey(key))
}

func (c *ServerContext) GetOrders(w web.ResponseWriter, r *web.Request) {
	orders, err := entities.ListOrders(c.Context)
	if err != nil {
		log.Errorf(c.Context, "Error fetching orders: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Unexpected error getting orders")
		return
	}

	c.ServeJson(http.StatusOK, orders)
}
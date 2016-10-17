package km

import (
	"github.com/gocraft/web"
	"entities"
	"google.golang.org/appengine/log"
	"net/http"
	"strconv"
	"google.golang.org/appengine/blobstore"
)

type AdminContext struct {
	*ServerContext
	User *entities.User
}

func (c *AdminContext) Auth(w web.ResponseWriter, r *web.Request, next web.NextMiddlewareFunc) {
	c.User = &entities.User{Email: "jcarm010@fiu.edu"}
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
	log.Infof(c.Context, "Modifying product [%s] with values: name[%s] price_cents[%s] quantity[%s] active[%s]", idStr, name, priceCentsStr, quantityStr, activeStr)
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
	err = entities.UpdateProduct(c.Context, product)
	if err != nil {
		log.Errorf(c.Context, "Error storing product: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Unexpected value storing product")
		return
	}
}

func (c *AdminContext) GetProductImageUploadUrl(w web.ResponseWriter, r *web.Request) {
	uploadURL, err := blobstore.UploadURL(c.Context, "/admin/product/image/upload", nil)
	if err != nil {
		log.Errorf(c.Context, "Failed to create upload url: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Failed to create upload url")
		return
	}

	log.Infof(c.Context, "Upload url: %+v", uploadURL)
	c.ServeJson(http.StatusOK, uploadURL)
}

func (c *AdminContext) ProductImageUploadHandler(w web.ResponseWriter, r *web.Request) {

}
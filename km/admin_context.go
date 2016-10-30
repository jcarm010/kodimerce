package km

import (
	"github.com/gocraft/web"
	"entities"
	"google.golang.org/appengine/log"
	"net/http"
	"strconv"
	"strings"
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
	picturesStr := r.FormValue("pictures")
	description := r.FormValue("description")
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
	log.Infof(c.Context, "Modifying category [%s] with values: name[%s] description[%s]", idStr, name, description)
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

	category := entities.NewCategory(name)
	category.Id = id
	category.Description = description
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
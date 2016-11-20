package km

import (
	"golang.org/x/net/context"
	"github.com/gocraft/web"
	"google.golang.org/appengine"
	"fmt"
	"encoding/json"
	"net/http"
	"regexp"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/appengine/log"
	"entities"
	"google.golang.org/appengine/datastore"
	"strings"
	"strconv"
	"html/template"
	"settings"
)

type ServerContext struct{
	Context context.Context
	w web.ResponseWriter
	r *web.Request
}

func (c *ServerContext) ServeJson(status int, value interface{}){
	c.w.Header().Add("Content-Type", "application/json")
	c.w.WriteHeader(status)
	bts, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(c.w, "%s", bts)
}

func (c *ServerContext) ServeHTML(status int, value interface{}){
	c.w.Header().Add("Content-Type", "text/html; charset=utf-8")
	c.w.WriteHeader(status)
	fmt.Fprintf(c.w, "%s", value)
}

func (c *ServerContext) ServeHTMLError(status int, value interface{}){
	var templates = template.Must(template.ParseGlob("views/template/*")) // cache this globally
	c.w.Header().Add("Content-Type", "text/html; charset=utf-8")
	c.w.WriteHeader(status)
	type ErrorView struct {
		Title string
		Message string
	}

	err := templates.ExecuteTemplate(c.w, "error-page", ErrorView {
		Title: settings.COMPANY_NAME + " | Error",
		Message: fmt.Sprintf("%s", value),
	})

	if err != nil {
		log.Errorf(c.Context, "Error parsing html file: %+v", err)
		fmt.Fprint(c.w, "Unexpected error, please try again later.")
		return
	}
}

func (c *ServerContext) InitServerContext(w web.ResponseWriter, r *web.Request, next web.NextMiddlewareFunc){
	c.Context = appengine.NewContext(r.Request)
	c.w = w
	c.r = r
	next(w, r)
}

func (c *ServerContext) RegisterUser(w web.ResponseWriter, r *web.Request){
	err := r.ParseForm()
	if err != nil {
		c.ServeJson(http.StatusBadRequest, "Could not read values.")
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")
	if email == "" {
		c.ServeJson(http.StatusBadRequest, "Missing email.")
		return
	}

	if password == "" {
		c.ServeJson(http.StatusBadRequest, "Missing password.")
		return
	}

	re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	if !re.MatchString(email) {
		c.ServeJson(http.StatusBadRequest, "Invalid email.")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Errorf(c.Context, "Error hashing password[%s]: %+v", password, err)
		c.ServeJson(http.StatusInternalServerError, "Unexpected error creating user.")
		return
	}

	user := entities.NewUser(email)
	user.PasswordHash = string(hashedPassword)
	err = entities.CreateUser(c.Context, user)
	if err == entities.ErrUserAlreadyExists {
		log.Errorf(c.Context, "User already exists: %s", email)
		c.ServeJson(http.StatusBadRequest, "User already exists.")
		return
	}else if(err != nil){
		log.Errorf(c.Context, "Error creating user[%s]: %+v", email, err)
		c.ServeJson(http.StatusInternalServerError, "Unexpected error creating user.")
		return
	}
}

func (c *ServerContext) LoginUser(w web.ResponseWriter, r *web.Request){
	err := r.ParseForm()
	if err != nil {
		c.ServeJson(http.StatusBadRequest, "Could not read values.")
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")
	if email == "" {
		c.ServeJson(http.StatusBadRequest, "Missing email.")
		return
	}

	if password == "" {
		c.ServeJson(http.StatusBadRequest, "Missing password.")
		return
	}

	re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	if !re.MatchString(email) {
		c.ServeJson(http.StatusBadRequest, "Invalid email.")
		return
	}

	user, err := entities.GetUser(c.Context, email)
	if err == datastore.ErrNoSuchEntity {
		log.Errorf(c.Context, "User not found: %s", email)
		c.ServeJson(http.StatusBadRequest, "User not found.")
		return
	}else if(err != nil){
		log.Errorf(c.Context, "Error getting user[%s]: %+v", email, err)
		c.ServeJson(http.StatusInternalServerError, "Unexpected error getting user.")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		log.Errorf(c.Context, "User passwords do not match: %+v", err)
		c.ServeJson(http.StatusBadRequest, "User not found.")
		return
	}

	userSession, err := entities.CreateUserSession(c.Context, email)
	if err != nil {
		log.Errorf(c.Context, "Error creating user session: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Unexpected error creating session.")
		return
	}

	cookie := &http.Cookie{Name: "km-session", Value: userSession.SessionToken, HttpOnly: false}
	http.SetCookie(w, cookie)

	if user.UserType == "admin" {
		c.ServeJson(http.StatusOK, "/admin")
		return
	}

	c.ServeJson(http.StatusOK, "/")
}

func (c *ServerContext) CreateOrder(w web.ResponseWriter, r *web.Request){
	log.Infof(c.Context, "Creating new order")
	err := r.ParseForm()
	if err != nil {
		log.Errorf(c.Context, "Error parsing form: %+v", err)
		c.ServeJson(http.StatusBadRequest, "Could not understand the request. Please try again later.")
		return
	}

	productIdsCommaStr := r.FormValue("product_ids")
	log.Infof(c.Context, "productIdsCommaStr[%+v]", productIdsCommaStr)
	productIdsStr := strings.Split(productIdsCommaStr, ",")
	if len(productIdsStr) < 1 || productIdsStr[0] == "" {
		log.Errorf(c.Context, "No products found: %+v", productIdsStr)
		c.ServeJson(http.StatusBadRequest, "Can't create an order without products. Please add products to your shopping cart and try again.")
		return
	}

	productIds := make([]int64, 0)
	for _, productIdStr := range productIdsStr {
		id, err := strconv.ParseInt(productIdStr, 10, 64)
		if err != nil {
			log.Errorf(c.Context, "Could not parse product id[%s]: %+v", productIdStr, err)
			continue
		}

		productIds = append(productIds, id)
	}

	product, err := entities.CreateOrder(c.Context, productIds)
	if err != nil {
		log.Errorf(c.Context, "Error creating product: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Could not create the order at this moment. Please try again later.")
		return
	}

	c.ServeJson(http.StatusOK, product)
}

func (c *ServerContext) UpdateOrder(w web.ResponseWriter, r *web.Request){
	err := r.ParseForm()
	if err != nil {
		log.Errorf(c.Context, "Error parsing form: %+v", err)
		c.ServeJson(http.StatusBadRequest, "Could not understand the request. Please try again later.")
		return
	}

	idStr := r.FormValue("id")
	shippingName := r.FormValue("shipping_name")
	shippingAddress := r.FormValue("shipping_address")
	email := r.FormValue("email")
	phone := r.FormValue("phone")
	checkoutStep := r.FormValue("checkout_step")

	log.Infof(c.Context, "Updating order idStr[%s] shippingName[%s] shippingAddress[%s] email[%s] phone[%s] checkoutStep[%s]", idStr, shippingName, shippingAddress, email, phone, checkoutStep)

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
	order.ShippingAddress = shippingAddress
	order.Email = email
	order.Phone = phone
	order.CheckoutStep = checkoutStep

	err = entities.UpdateOrder(c.Context, order)
	if err != nil {
		log.Errorf(c.Context, "Error updating order: %+v", err)
		c.ServeJson(http.StatusBadRequest, "Could not update order. Please try again later.")
		return
	}

	c.ServeJson(http.StatusOK, "")
}
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
	"github.com/jcarm010/kodimerce/entities"
	"google.golang.org/appengine/datastore"
	"strings"
	"strconv"
	"html/template"
	"github.com/jcarm010/kodimerce/settings"
	"github.com/jcarm010/kodimerce/paypal"
	"github.com/jcarm010/kodimerce/smartyaddress"
	"github.com/jcarm010/kodimerce/emailer"
	"bytes"
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
	var templates = template.Must(template.ParseGlob("views/templates/*")) // cache this globally
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
	}else if err != nil{
		log.Errorf(c.Context, "Error creating user[%s]: %+v", email, err)
		c.ServeJson(http.StatusInternalServerError, "Unexpected error creating user.")
		return
	}

	err = emailer.SendEmail(
		c.Context,
		fmt.Sprintf("%s<%s>", settings.COMPANY_NAME, settings.EMAIL_SENDER),
		user.Email,
		fmt.Sprintf("Welcome to %s", settings.COMPANY_NAME),
		fmt.Sprintf("Thank you for registering to %s", settings.COMPANY_NAME),
		[]string{},
	)

	if err != nil {
		log.Errorf(c.Context, "Couldn't send email: %v", err)
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

	itemsStr := r.FormValue("products")
	log.Infof(c.Context, "itemsStr: %s", itemsStr)
	orderProducts := make([]*entities.OrderProduct, 0)
	err = json.Unmarshal([]byte(itemsStr), &orderProducts)
	if err != nil {
		log.Errorf(c.Context, "Error reading products: %+v", err)
		c.ServeJson(http.StatusBadRequest, "Could not find products.")
		return
	}

	log.Infof(c.Context, "Products received: %+v", orderProducts)
	quantities := make([]int64, 0)
	productIds := make([]int64, 0)
	productDetails := make([]*entities.ProductDetails, 0)
	for _, product := range orderProducts {
		productId := product.Id
		productIds = append(productIds, productId)
		quantities = append(quantities, product.Quantity)
		productDetails = append(productDetails, &entities.ProductDetails{
			ProductId: product.Id,
			Time: product.Time,
			Date: product.Date,
		})
	}

	log.Infof(c.Context, "Creating order with ProductIds: %+v and Quantities: %+v", productIds, quantities)
	products, err := entities.GetProducts(c.Context, productIds)
	if err != nil {
		log.Errorf(c.Context, "Error getting products: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Could not create the order at this moment. Please try again later.")
		return
	}

	order, err := entities.CreateOrder(c.Context, products, quantities, productDetails)
	if err != nil {
		log.Errorf(c.Context, "Error creating order: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Could not create the order at this moment. Please try again later.")
		return
	}

	c.ServeJson(http.StatusOK, order)
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

	if order.Status != entities.ORDER_STATUS_STARTED {
		log.Errorf(c.Context, "Order is not in started status[%+v]: %+v", order, err)
		c.ServeJson(http.StatusBadRequest, "Order has already been placed.")
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

func (c *ServerContext) CheckOrderAddress(w web.ResponseWriter, r *web.Request){
	err := r.ParseForm()
	if err != nil {
		log.Errorf(c.Context, "Error parsing form: %+v", err)
		c.ServeJson(http.StatusBadRequest, "Could not understand the request. Please try again later.")
		return
	}

	idStr := r.FormValue("id")
	shippingLine1 := r.FormValue("shipping_line_1")
	shippingLine2 := r.FormValue("shipping_line_2")
	city := r.FormValue("city")
	state := r.FormValue("state")
	postalCode := r.FormValue("postal_code")
	countryCode := r.FormValue("country_code")

	log.Infof(c.Context, "Verifying order address idStr[%s] shippingLine1[%s] shippingLine2[%s] city[%s] state[%s] postalCode[%s] countryCode[%s]",
		idStr, shippingLine1, shippingLine2, city, state, postalCode, countryCode)

	if shippingLine1 == "" {
		log.Errorf(c.Context, "Missing shipping line 1")
		c.ServeJson(http.StatusBadRequest, "Missing shipping address")
		return
	}

	if city == "" {
		log.Errorf(c.Context, "Missing city")
		c.ServeJson(http.StatusBadRequest, "Missing city")
		return
	}

	if state == "" {
		log.Errorf(c.Context, "Missing state")
		c.ServeJson(http.StatusBadRequest, "Missing state")
		return
	}

	if postalCode == "" {
		log.Errorf(c.Context, "Missing postal code")
		c.ServeJson(http.StatusBadRequest, "Missing postal code")
		return
	}

	if countryCode == "" {
		log.Errorf(c.Context, "Missing country code")
		c.ServeJson(http.StatusBadRequest, "Missing country code")
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

	if order.Status != entities.ORDER_STATUS_STARTED {
		log.Errorf(c.Context, "Order is not in started status[%+v]: %+v", order, err)
		c.ServeJson(http.StatusBadRequest, "Order has already been placed.")
		return
	}

	lookup := &smartyaddress.Lookup{
		Street: shippingLine1,
		Street2: shippingLine2,
		City: city,
		State: state,
		ZIPCode: postalCode,
	}

	candidate, err := smartyaddress.CheckUSAddress(c.Context, lookup)
	if err == smartyaddress.ADDRESS_NOT_FOUND_ERROR {
		log.Errorf(c.Context, "Could not find address address: %+v", err)
		c.ServeJson(http.StatusBadRequest, "Could not find your address. Please review your address for errors or mispellings.")
		return
	}

	if err != nil {
		log.Errorf(c.Context, "Error looking up address: %+v", err)
		c.ServeJson(http.StatusBadRequest, "Could not find address. Please try again later.")
		return
	}

	c.ServeJson(http.StatusOK, candidate)
}

func (c *ServerContext) CreatePaypalPayment(w web.ResponseWriter, r *web.Request){
	type CreatePaymentResponse struct {
		Error string `json:"error"`
		PaymentID string `json:"paymentID"`
	}

	response := CreatePaymentResponse {}

	orderIdStr := r.URL.Query().Get("order")
	if orderIdStr == "" {
		log.Errorf(c.Context, "Missing order")
		response.Error = "Missing order"
		c.ServeJson(http.StatusBadRequest, response)
		return
	}

	orderId, err := strconv.ParseInt(orderIdStr, 10, 64)
	if err != nil {
		log.Errorf(c.Context, "Error parsing orderId: %+v", err)
		response.Error = "Invalid order id"
		c.ServeJson(http.StatusBadRequest, response)
		return
	}

	log.Infof(c.Context, "orderIdStr: %v", orderId)
	order, err := entities.GetOrder(c.Context, orderId)
	if err != nil {
		log.Errorf(c.Context, "Error getting order: %+v", err)
		response.Error = "Error finding order"
		c.ServeJson(http.StatusBadRequest, response)
		return
	}

	if order.Status != entities.ORDER_STATUS_STARTED {
		log.Errorf(c.Context, "Order is not in started status[%+v]: %+v", order, err)
		response.Error = "Order has already been placed."
		c.ServeJson(http.StatusBadRequest, response)
		return
	}

	log.Infof(c.Context, "Order: %+v", order)
	id, err := paypal.CreatePayment(c.Context, order)
	if err != nil {
		log.Errorf(c.Context, "Error creating paypal payment: %+v", err)
		response.Error = "Unexpected error creating paypal payment"
		c.ServeJson(http.StatusInternalServerError, response)
		return
	}

	order.PaypalPaymentId = id
	err = entities.UpdateOrder(c.Context, order)
	if err != nil {
		log.Errorf(c.Context, "Error storing paypal payment id: %+v", err)
		response.Error = "Unexpected error creating paypal payment"
		c.ServeJson(http.StatusInternalServerError, response)
		return
	}

	response.PaymentID = id
	c.ServeJson(http.StatusOK, response)
}

func (c *ServerContext) ExecutePaypalPayment(w web.ResponseWriter, r *web.Request){
	log.Infof(c.Context, "Executing Paypal payment....")
	err := r.ParseForm()
	if err != nil {
		log.Errorf(c.Context, "Error parsing parameters: %+v", err)
		c.ServeJson(http.StatusBadRequest, "Invalid parameters")
		return
	}

	idStr := r.FormValue("id")
	if idStr == "" {
		c.ServeJson(http.StatusBadRequest, "Missing order id")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Errorf(c.Context, "Error parsing order id[%s]: %+v", idStr, err)
		c.ServeJson(http.StatusBadRequest, "Invalid order id")
		return
	}

	log.Infof(c.Context, "Confirming order id: %v", id)
	order, err := entities.GetOrder(c.Context, id)
	if err != nil {
		log.Errorf(c.Context, "Error getting order id: %+v", err)
		c.ServeJson(http.StatusBadRequest, "Could not find order")
		return
	}

	if order.Status != entities.ORDER_STATUS_STARTED {
		log.Errorf(c.Context, "Order is not in started status [%+v]", order)
		c.ServeJson(http.StatusBadRequest, "Order has already been placed.")
		return
	}

	err = paypal.ExecutePayment(c.Context, order)
	if err != nil {
		log.Errorf(c.Context, "Error executing payment: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Unexpecting error executing payment")
		return
	}

	order.Status = entities.ORDER_STATUS_PENDING
	err = entities.UpdateOrder(c.Context, order)
	if err != nil {
		log.Errorf(c.Context, "Error updating order status: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Unexpecting error executing payment")
		return
	}

	for _, productId := range order.ProductIds {
		err = entities.DecreaseProductInventory(c.Context, productId, 1)
		if err != nil {
			log.Errorf(c.Context, "Error decresing inventory for productId[%v]: %+v", productId, err)
		}
	}

	proto := "http"
	if r.Request.TLS != nil {
		proto = "https"
	}
	serverRoot := fmt.Sprintf("%s://%s", proto, r.Host)
	confirmationUrl := fmt.Sprintf("%s/order?id=%v", serverRoot, order.Id)

	var templates = template.Must(template.ParseGlob("emailer/templates/*")) // cache this globally
	type OrderConfirmationEmail struct {
		CompanyName string
		ConfirmationUrl string
		HostRoot string
		ContactEmail string
	}
	var doc bytes.Buffer
	err = templates.ExecuteTemplate(&doc, "email-order", OrderConfirmationEmail{
		CompanyName: settings.COMPANY_NAME,
		ConfirmationUrl: confirmationUrl,
		HostRoot: serverRoot,
		ContactEmail: settings.COMPANY_SUPPORT_EMAIL,
	})

	if err != nil {
		log.Errorf(c.Context, "Error parsing email template: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Unexpected Error, please try again later.")
		return
	}

	//send a notification email to the buyer
	err = emailer.SendEmail(
		c.Context,
		fmt.Sprintf("%s<%s>", settings.COMPANY_NAME, settings.EMAIL_SENDER),
		order.Email,
		"Order Confirmation",
		doc.String(),
		[]string{settings.COMPANY_ORDERS_EMAIL},
	)

	if err != nil {
		log.Errorf(c.Context, "Couldn't send email: %v", err)
	}
}

func (c *ServerContext) GetProducts(w web.ResponseWriter, r *web.Request){
	idsStrRaw := r.FormValue("ids")
	idsStr := strings.Split(idsStrRaw, ",")
	log.Infof(c.Context, "Gettting products: %s", idsStr)
	if len(idsStr) == 1 && idsStr[0] == "" {
		c.ServeJson(http.StatusBadRequest, "Missing product ids")
		return
	}

	ids := make([]int64, len(idsStr))
	for index, idStr := range idsStr {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.Errorf(c.Context, "Could not parse product id[%s]: %s", idStr, err)
			c.ServeJson(http.StatusBadRequest, "Invalid product id")
			return
		}

		ids[index] = id
	}

	products, err := entities.GetProducts(c.Context, ids)
	if err != nil {
		log.Errorf(c.Context, "Error getting products: %s", err)
		c.ServeJson(http.StatusInternalServerError, "Unexpected error getting products")
		return
	}

	c.ServeJson(http.StatusOK, products)
}

func (c *ServerContext) PostContactMessage(w web.ResponseWriter, r *web.Request){
	err := r.ParseForm()
	if err != nil {
		log.Errorf(c.Context, "Error parsing request: %s", err)
		c.ServeJson(http.StatusBadRequest, "Unexpected error parsing request")
		return
	}

	q := r.Form
	name := q.Get("name")
	if name == "" {
		log.Errorf(c.Context, "Missing name")
		c.ServeJson(http.StatusBadRequest, "Please provide a name.")
		return
	}

	email := q.Get("email")
	if email == "" {
		log.Errorf(c.Context, "Missing email")
		c.ServeJson(http.StatusBadRequest, "Please provide an email so that we can get back to you.")
		return
	}

	message := q.Get("message")
	if message == "" {
		log.Errorf(c.Context, "Missing message")
		c.ServeJson(http.StatusBadRequest, "Please provide a message so that we can address your questions or concerns.")
		return
	}

	log.Infof(c.Context, "Sending message name[%s] email[%s] message[%s]", name, email, message)
	body := fmt.Sprintf("Customer %s (%s) has sent you a message: %s", name, email, message)
	err = emailer.SendEmail(
		c.Context,
		fmt.Sprintf("%s<%s>", settings.COMPANY_NAME, settings.EMAIL_SENDER),
		settings.COMPANY_SUPPORT_EMAIL,
		"Customer Message",
		body,
		[]string{settings.COMPANY_ORDERS_EMAIL},
	)

	if err != nil {
		log.Errorf(c.Context, "Error Sending email: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Could not send message at this time. Please try again later.")
		return
	}

	c.ServeJson(http.StatusOK, "")
}
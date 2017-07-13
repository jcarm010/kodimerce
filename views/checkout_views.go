package views

import (
	"github.com/gocraft/web"
	"github.com/jcarm010/kodimerce/km"
	"google.golang.org/appengine/log"
	"net/http"
	"github.com/jcarm010/kodimerce/settings"
	"github.com/dustin/gojson"
	"strconv"
	"github.com/jcarm010/kodimerce/entities"
	"github.com/jcarm010/kodimerce/view"
)

type CheckoutStep struct {
	Name string `json:"name"`
	Label string `json:"label"`
	Number int `json:"number"`
	Current bool `json:"current"`
	Component string `json:"component"`
}

func (c *CheckoutStep) HasName (name string) bool {
	return c.Name == name
}

func (c *CheckoutStep) String () string {
	bts, err := json.Marshal(c)
	if err != nil {
		return err.Error()
	}

	return string(bts)
}

func RenderCheckoutView(c *km.ServerContext, w web.ResponseWriter, r *web.Request) {
	orderIdStr := r.URL.Query().Get("order")
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

	stepName := r.PathParams["step"]
	if stepName == "" {
		stepName = "shipinfo"
	}

	log.Infof(c.Context, "Rendering orderId[%v] step[%s] order[%+v]", orderId, stepName, order)


	checkoutSteps := []*CheckoutStep{
		{Name: "shipinfo", Label:"Shipping Information", Number: 1, Current: true, Component: "km-checkout-shipinfo"},
		{Name: "payinfo", Label:"Payment Information", Number: 2, Current: false, Component: "km-checkout-payinfo"},
		{Name: "confirm", Label:"Review and Confirm", Number: 3, Current: false, Component: "km-checkout-confirm"},
	}

	currentStep := checkoutSteps[0]
	nextStep := checkoutSteps[0]

	for index, step := range checkoutSteps {
		current := step.Name == stepName
		step.Current = current
		if current {
			currentStep = step
		}

		if current && index < len(checkoutSteps) - 1 {
			nextStep = checkoutSteps[index + 1]
		}
	}

	err = km.Templates.ExecuteTemplate(w, "checkout-page", struct{
		*view.View
		CheckoutSteps []*CheckoutStep `json:"checkout_steps"`
		CurrentStep *CheckoutStep `json:"current_step"`
		NextStep *CheckoutStep `json:"next_step"`
		Order *entities.Order `json:"order"`
		PaypalEnvironment string `json:"paypal_environment"`
	}{
		View: c.NewView("Checkout | " + settings.COMPANY_NAME, ""),
		CheckoutSteps:checkoutSteps,
		CurrentStep:currentStep,
		NextStep: nextStep,
		Order: order,
		PaypalEnvironment: settings.PAYPAL_ENVIRONMENT,
	})

	if err != nil {
		log.Errorf(c.Context, "Error parsing html file: %+v", err)
		c.ServeHTMLError(http.StatusInternalServerError, "Unexpected error, please try again later.")
		return
	}
}

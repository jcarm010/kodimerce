package settings

import (
	"os"
	"net/http"
	"fmt"
	"strconv"
)

var (
	COMPANY_NAME = os.Getenv("COMPANY_NAME")
	COMPANY_SUPPORT_EMAIL = os.Getenv("COMPANY_SUPPORT_EMAIL")
	COMPANY_ORDERS_EMAIL = os.Getenv("COMPANY_ORDERS_EMAIL")
	TAX_PERCENT float64 = 0.0
	COMPANY_URL = os.Getenv("COMPANY_URL")
	PAYPAL_ENVIRONMENT = os.Getenv("PAYPAL_ENVIRONMENT")
	PAYPAL_API_URL = os.Getenv("PAYPAL_API_URL")
	PAYPAL_EMAIL = os.Getenv("PAYPAL_EMAIL")
	PAYPAL_ACCOUNT = os.Getenv("PAYPAL_ACCOUNT")
	PAYPAL_API_CLIENT_ID = os.Getenv("PAYPAL_API_CLIENT_ID")
	PAYPAL_API_CLIENT_SECRET = os.Getenv("PAYPAL_API_CLIENT_SECRET")
	PAYPAL_ALLOWED_PAYMENT_OPTION = os.Getenv("PAYPAL_ALLOWED_PAYMENT_OPTION") //posible: UNRESTRICTED, INSTANT_FUNDING_SOURCE, IMMEDIATE_PAY
	PAYPAL_NOTE_TO_PAYER = os.Getenv("PAYPAL_NOTE_TO_PAYER")
	SMARTYSTREETS_AUTH_ID = os.Getenv("SMARTYSTREETS_AUTH_ID")
	SMARTYSTREETS_AUTH_TOKEN = os.Getenv("SMARTYSTREETS_AUTH_TOKEN")
	EMAIL_SENDER = os.Getenv("EMAIL_SENDER")
	SENDGRID_KEY = os.Getenv("SENDGRID_KEY")
)

func init() {
	TAX_PERCENT, _ = strconv.ParseFloat(os.Getenv("TAX_PERCENT"), 64)
}

func ServerUrl(r *http.Request) string {
	httpHeader := "http"
	if r.TLS != nil {
		httpHeader = "https"
	}

	return fmt.Sprintf("%s://%s", httpHeader, r.Host)
}
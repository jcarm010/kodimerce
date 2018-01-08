package paypal

import (
	"golang.org/x/net/context"
	"path"
	"github.com/jcarm010/kodimerce/settings"
	"encoding/json"
	"google.golang.org/appengine/log"
	"github.com/jcarm010/kodimerce/entities"
	"fmt"
	"bytes"
	"errors"
	"io/ioutil"
	"net/url"
	"net/http"
	"google.golang.org/appengine/urlfetch"
	"runtime"
	"crypto/tls"
	"time"
)

type PaypalCreatePaymentRequest struct {
	Intent string `json:"intent"`
	Payer map[string]string `json:"payer"`
	Transactions []*Transaction `json:"transactions"`
	NoteToPayer string `json:"note_to_payer"`
	RedirectUrls *RedirectUrls `json:"redirect_urls"`
}

type PaypalExecutePaymentRequest struct {
	PayerId string `json:"payer_id"`
}

type PaypalCreatePaymentResponse struct {
	Id string `json:"id"`
}

type RedirectUrls struct {
	ReturnUrl string `json:"return_url"`
	CancelUrl string `json:"cancel_url"`
}

func getAccessToken(ctx context.Context) (string, error) {
	generalSettings := settings.GetGlobalSettings(ctx)
	u, err := url.Parse(generalSettings.PayPalApiUrl)
	if err != nil {
		return "", err
	}

	u.Path = path.Join(u.Path, "oauth2/token")
	oauthUrl := u.String()
	log.Infof(ctx, "OAuth Url: %s", oauthUrl)

	req, err := http.NewRequest(http.MethodPost, oauthUrl, bytes.NewBuffer([]byte("grant_type=client_credentials")))
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en_US")
	req.SetBasicAuth(generalSettings.PayPalApiClientId, generalSettings.PayPalApiClientSecret)

	client := getClient(ctx)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	bts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf("Bad status on oath response[%s]: %s", resp.Status, bts))
	}

	log.Infof(ctx, "PayPal access token response: %s", bts)
	type TokenResponse struct {
		AccessToken string `json:"access_token"`
	}

	token := &TokenResponse{}
	err = json.Unmarshal(bts, token)
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}

func getClient (ctx context.Context) *http.Client {
	ctx_with_deadline, _ := context.WithTimeout(ctx, time.Duration(40) * time.Second)
	client := &http.Client{
		Transport:&urlfetch.Transport{Context: ctx_with_deadline},
	}

	if runtime.GOOS == "darwin" {
		/* Mac have an issue openssl issue dealing with TLS.
		This should only occur on development environments. */
		tlsconf := &tls.Config{MinVersion: tls.VersionTLS10}
		tr := &http.Transport{
			TLSClientConfig: tlsconf,
			DisableCompression: true,
		}

		client.Transport = tr
	}

	return client
}

func CreatePayment(ctx context.Context, order *entities.Order, companyUrl string) (string, error) {
	products := order.Products
	log.Infof(ctx, "Products: %+v", products)
	items := make([]*Item, len(products))
	var subtotalCents int64 = 0
	var taxCents int64 = 0
	var shippingCents int64 = 0
	var handlingFeeCents int64 = 0
	var shippingDiscountCents int64 = 0
	var insuranceCents int64 = 0

	for index, product := range products {
		var qty int64 = 0
		if order.Quantities != nil &&  len(order.Quantities) > 0 {
			qty = order.Quantities[index]
		}

		if qty == 0 {
			qty = 1
		}

		productDetails := order.ProductDetails[index]
		var name string
		var priceCents int64
		if product.HasPricingOptions {
			name = fmt.Sprintf("%s - %s", product.Name, productDetails.PricingOption.Label)
			priceCents += productDetails.PricingOption.PriceCents
		} else {
			name = product.Name
			priceCents = product.GetPriceCents()
		}

		subtotalCents += priceCents * qty
		u, err := url.Parse(companyUrl)
		if err != nil {
			return "", err
		}

		u.Path = path.Join(u.Path, fmt.Sprintf("product/%v", product.Id))
		productUrl := u.String()
		items[index] = NewItem(fmt.Sprintf("%v", product.Id), name, string(product.Description), int(qty), priceCents, 0, productUrl)
	}

	taxCents = int64(float64(subtotalCents) * order.TaxPercent / 100)
	amount := NewAmount(subtotalCents, taxCents, shippingCents, handlingFeeCents, shippingDiscountCents, insuranceCents)

	var shippingAddress *ShippingAddress = nil
	if !order.NoShipping {
		shippingAddress = &ShippingAddress{
			Line1: order.ShippingLine1,
			Line2: order.ShippingLine2,
			City: order.City,
			CountryCode: order.CountryCode,
			PostalCode: order.PostalCode,
			State: order.State,
			Phone: order.Phone,
			RecipientName: order.ShippingName,
		}
	}

	globalSettings := settings.GetGlobalSettings(ctx)
	transaction := NewTransaction(
		fmt.Sprintf("%v",order.Id),
		fmt.Sprintf("An order from %s.", globalSettings.CompanyName),
		amount,
		items,
		shippingAddress,
		globalSettings.PayPalAllowedPaymentOption,
	)

	u, err := url.Parse(companyUrl)
	if err != nil {
		return "", err
	}

	u.Path = path.Join(u.Path, fmt.Sprintf("paypal/return/%v", order.Id))
	returnUrl := u.String()

	u, err = url.Parse(companyUrl)
	if err != nil {
		return "", err
	}

	u.Path = path.Join(u.Path, fmt.Sprintf("paypal/cancel/%v", order.Id))
	cancelUrl := u.String()

	paypalRequest := PaypalCreatePaymentRequest{
		Intent: "sale",
		Payer: map[string]string{"payment_method":"paypal"},
		NoteToPayer: globalSettings.PayPalNoteToPayer,
		Transactions: []*Transaction{transaction},
		RedirectUrls: &RedirectUrls{
			ReturnUrl: returnUrl,
			CancelUrl: cancelUrl,
		},
	}

	jsonStr, err := json.Marshal(paypalRequest)
	if err != nil {
		return "", err
	}

	log.Infof(ctx, "Making paypal payment create request: %s", jsonStr)
	u, err = url.Parse(globalSettings.PayPalApiUrl)
	if err != nil {
		return "", err
	}

	u.Path = path.Join(u.Path, "payments/payment")
	paypalUrl := u.String()
	log.Debugf(ctx, "Paypal url: %s", paypalUrl)
	req, err := http.NewRequest(http.MethodPost, paypalUrl, bytes.NewBuffer(jsonStr))
	if err != nil {
		return "", err
	}

	accessToken, err := getAccessToken(ctx)
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	client := getClient(ctx)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	bts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	log.Debugf(ctx, "Paypal response: %s", bts)

	if resp.StatusCode != 201 {
		return "", errors.New(fmt.Sprintf("Paypal responded with status[%s]: %s", resp.Status, bts))
	}

	r := &PaypalCreatePaymentResponse{}
	err = json.Unmarshal(bts, r)
	if err != nil {
		return "", err
	}

	return r.Id, nil
}

func ExecutePayment(ctx context.Context, order *entities.Order) (error) {
	executeRequest := PaypalExecutePaymentRequest{
		PayerId: order.PaypalPayerId,
	}

	jsonStr, err := json.Marshal(executeRequest)
	if err != nil {
		return err
	}

	globalSettings := settings.GetGlobalSettings(ctx)
	u, err := url.Parse(globalSettings.PayPalApiUrl)
	if err != nil {
		return err
	}

	u.Path = path.Join(u.Path, "payments/payment/" + order.PaypalPaymentId +"/execute")
	paypalUrl := u.String()
	log.Debugf(ctx, "Paypal url: %s", paypalUrl)
	req, err := http.NewRequest(http.MethodPost, paypalUrl, bytes.NewBuffer(jsonStr))
	if err != nil {
		return err
	}

	accessToken, err := getAccessToken(ctx)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	client := getClient(ctx)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	bts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.Debugf(ctx, "Paypal response: %s", bts)
	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Paypal responded with status[%s]: %s", resp.Status, bts))
	}

	return nil
}
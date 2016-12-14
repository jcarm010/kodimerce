package paypal

import (
	"golang.org/x/net/context"
	"path"
	"settings"
	"encoding/json"
	"google.golang.org/appengine/log"
	"entities"
	"fmt"
	"bytes"
	"errors"
	"io/ioutil"
	"net/url"
	"net/http"
	"google.golang.org/appengine/urlfetch"
	"runtime"
	"crypto/tls"
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
	u, err := url.Parse(settings.PAYPAL_API_URL)
	if err != nil {
		return "", err
	}

	u.Path = path.Join(u.Path, "oauth2/token")
	url := u.String()
	log.Infof(ctx, "OAuth Url: %s", url)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte("grant_type=client_credentials")))
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en_US")
	req.SetBasicAuth(settings.PAYPAL_API_CLIENT_ID, settings.PAYPAL_API_CLIENT_SECRET)

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
	client := urlfetch.Client(ctx)
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

func CreatePayment(ctx context.Context, order *entities.Order) (string, error) {

	products, err := entities.GetProducts(ctx, order.ProductIds)
	if err != nil {
		return "", err
	}

	log.Infof(ctx, "Products: %+v", products)
	items := make([]*Item, len(products))
	var subtotalCents int64 = 0
	var taxCents int64 = 0
	var shippingCents int64 = 0
	var handlingFeeCents int64 = 0
	var shippingDiscountCents int64 = 0
	var insuranceCents int64 = 0

	for index, product := range products {
		subtotalCents += product.PriceCents
		u, err := url.Parse(settings.COMPANY_URL)
		if err != nil {
			return "", err
		}

		u.Path = path.Join(u.Path, fmt.Sprintf("product/%v",product.Id))
		productUrl := u.String()
		items[index] = NewItem(fmt.Sprintf("%v", product.Id), product.Name, product.Description, 1, product.PriceCents, 0, productUrl)
	}

	amount := NewAmount(subtotalCents, taxCents, shippingCents, handlingFeeCents, shippingDiscountCents, insuranceCents)

	transaction := NewTransaction(
		fmt.Sprintf("%v",order.Id),
		"An order of unique fashion pieces.",
		amount,
		items,
		&ShippingAddress{
			Line1: order.ShippingLine1,
			Line2: order.ShippingLine2,
			City: order.City,
			CountryCode: order.CountryCode,
			PostalCode: order.PostalCode,
			State: order.State,
			Phone: order.Phone,
			RecipientName: order.ShippingName,
		},
	)

	u, err := url.Parse(settings.COMPANY_URL)
	if err != nil {
		return "", err
	}

	u.Path = path.Join(u.Path, fmt.Sprintf("paypal/return/%v", order.Id))
	returnUrl := u.String()

	u, err = url.Parse(settings.COMPANY_URL)
	if err != nil {
		return "", err
	}

	u.Path = path.Join(u.Path, fmt.Sprintf("paypal/cancel/%v", order.Id))
	cancelUrl := u.String()

	paypalRequest := PaypalCreatePaymentRequest{
		Intent: "sale",
		Payer: map[string]string{"payment_method":"paypal"},
		NoteToPayer: settings.PAYPAL_NOTE_TO_PAYER,
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
	u, err = url.Parse(settings.PAYPAL_API_URL)
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

	u, err := url.Parse(settings.PAYPAL_API_URL)
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
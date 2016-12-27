package entities

import (
	"time"
	"google.golang.org/appengine/datastore"
	"golang.org/x/net/context"
	"github.com/dustin/gojson"
	"strings"
)

const (
	ENTITY_ORDER = "order"
	ORDER_STATUS_STARTED = "started"
	ORDER_STATUS_PENDING = "pending"
)

type Order struct {
	Id int64 `datastore:"-" json:"id"`
	ShippingName string `datastore:"shipping_name" json:"shipping_name"`
	ShippingLine1 string `datastore:"shipping_line_1,noindex" json:"shipping_line_1"`
	ShippingLine2 string `datastore:"shipping_line_2,noindex" json:"shipping_line_2"`
	City string `datastore:"city" json:"city"`
	State string `datastore:"state" json:"state"`
	PostalCode string `datastore:"postal_code" json:"postal_code"`
	CountryCode string `datastore:"country_code" json:"country_code"`
	Email string `datastore:"email" json:"email"`
	Phone string `datastore:"phone" json:"phone"`
	ProductIds []int64 `datastore:"product_ids,noindex" json:"product_ids"`
	Status string `datastore:"status" json:"status"`
	CheckoutStep string `datastore:"checkout_step" json:"checkout_step"`
	Created time.Time `datastore:"created" json:"created"`
	PaypalPaymentId string `datastore:"paypal_payment_id" json:"paypal_payment_id"`
	PaypalPayerId string `datastore:"paypal_payer_id" json:"paypal_payer_id"`
	AddressVerified bool `datastore:"address_verified" json:"address_verified"`
}

func (o *Order) StatusCapitalized() string {
	return strings.ToUpper(o.Status)
}

func (o *Order) String() string {
	bts, _ := json.Marshal(o)
	return string(bts)
}

func NewOrder() *Order {
	return &Order{
		Created: time.Now(),
		CheckoutStep: "shipinfo",
		Status: ORDER_STATUS_STARTED,
	}
}

func CreateOrder(ctx context.Context, productIds []int64) (*Order, error) {
	order := NewOrder()
	order.ProductIds = productIds
	key, err := datastore.Put(ctx, datastore.NewIncompleteKey(ctx, ENTITY_ORDER, nil), order)
	if err != nil {
		return nil, err
	}

	order.Id = key.IntID()
	return order, nil
}

func GetOrder(ctx context.Context, orderId int64) (*Order, error) {
	order := &Order{}
	err := datastore.Get(ctx, datastore.NewKey(ctx, ENTITY_ORDER, "", orderId, nil), order)
	if err != nil {
		return nil, err
	}

	order.Id = orderId
	return order, nil
}

func UpdateOrder(ctx context.Context, order *Order) (error) {
	_, err := datastore.Put(ctx, datastore.NewKey(ctx, ENTITY_ORDER, "", order.Id, nil), order)
	if err != nil {
		return err
	}

	return nil
}
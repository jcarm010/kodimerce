package entities

import (
	"time"
	"google.golang.org/appengine/datastore"
	"golang.org/x/net/context"
)

const (
	ENTITY_ORDER = "order"
	ORDER_STATUS_STARTED = "started"
	ORDER_STATUS_PENDING = "pending"
)

type Order struct {
	Id int64 `datastore:"-" json:"id"`
	ShippingName string `datastore:"shipping_name" json:"shipping_name"`
	ShippingAddress string `datastore:"shipping_address,noindex" json:"shipping_address"`
	Email string `datastore:"email" json:"email"`
	Phone string `datastore:"phone" json:"phone"`
	ProductIds []int64 `datastore:"product_ids,noindex" json:"product_ids"`
	Status string `datastore:"status" json:"status"`
	CheckoutStep string `datastore:"checkout_step" json:"checkout_step"`
	Created time.Time `datastore:"created" json:"created"`
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
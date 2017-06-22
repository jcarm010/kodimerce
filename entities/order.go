package entities

import (
	"time"
	"google.golang.org/appengine/datastore"
	"golang.org/x/net/context"
	"github.com/dustin/gojson"
	"strings"
	"github.com/jcarm010/kodimerce/settings"
)

const (
	ENTITY_ORDER = "order"
	ORDER_STATUS_STARTED = "started"
	ORDER_STATUS_PENDING = "pending"
	ORDER_STATUS_PROCESSING = "processing"
	ORDER_STATUS_SHIPPED = "shipped"
	ORDER_STATUS_PROCESSED = "processed"
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
	Quantities []int64 `datastore:"quantities,noindex" json:"quantities"`
	Status string `datastore:"status" json:"status"`
	CheckoutStep string `datastore:"checkout_step" json:"checkout_step"`
	Created time.Time `datastore:"created" json:"created"`
	PaypalPaymentId string `datastore:"paypal_payment_id" json:"paypal_payment_id"`
	PaypalPayerId string `datastore:"paypal_payer_id" json:"paypal_payer_id"`
	AddressVerified bool `datastore:"address_verified" json:"address_verified"`
	Products []*Product `datastore:"-" json:"products"`
	ProductsSerial []byte `datastore:"products_serial" json:"-"`
	NoShipping bool `datastore:"no_shipping" json:"no_shipping"`
	ProductDetails []*ProductDetails `datastore:"-" json:"product_details"`
	TaxPercent float64 `datastore:"tax_percent" json:"tax_percent"`
}

func (o *Order) Load(ps []datastore.Property) error {
	datastore.LoadStruct(o, ps)//todo: should probably do something about this error?
	for _, ps := range ps {
		if ps.Name == "product_details" {
			valueBts, ok := ps.Value.([]byte)
			if !ok {
				continue
			}

			err := json.Unmarshal(valueBts, &o.ProductDetails)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (o *Order) Save() ([]datastore.Property, error) {
	productDetailsBts, err := json.Marshal(o.ProductDetails)
	if err != nil {
		return nil, err
	}

	properties, err := datastore.SaveStruct(o)
	if err != nil {
		return nil, err
	}

	properties = append(properties, datastore.Property{
		Name:  "product_details",
		Value: productDetailsBts,
		NoIndex: true,
	})

	return properties, nil
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

type ProductDetails struct {
	ProductId int64 `datastore:"product_id" json:"product_id"`
	Date string `datastore:"date" json:"date"`
	Time AvailableTime `datastore:"time" json:"time"`
}

type OrderProduct struct {
	*Product
	Quantity int64 `json:"quantity"`
	Date string `json:"date"`
	Time AvailableTime `json:"time"`
}


func (o *OrderProduct) String() string {
	bts, _ := json.Marshal(o)
	return string(bts)
}

func CreateOrder(ctx context.Context, products []*Product, quantities []int64, productDetails []*ProductDetails) (*Order, error) {
	noShipping := true
	for _, product := range products {
		if !product.NoShipping {
			noShipping = false
			break
		}
	}

	order := NewOrder()
	productIds := make([]int64, len(products))
	for i, product := range products {
		productIds[i] = product.Id
	}

	order.ProductIds = productIds
	order.Products = products
	order.Quantities = quantities
	order.NoShipping = noShipping
	order.TaxPercent = settings.TAX_PERCENT
	order.ProductDetails = productDetails
	bts, err := json.Marshal(products)
	if err != nil {
		return nil, err
	}

	order.ProductsSerial = bts
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

	products := make([]*Product, 0)
	err = json.Unmarshal(order.ProductsSerial, &products)
	if err != nil {
		return nil, err
	}

	order.Products = products
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

func ListOrders(ctx context.Context) ([]*Order, error) {
	orders := make([]*Order, 0)
	keys, err := datastore.NewQuery(ENTITY_ORDER).GetAll(ctx, &orders)
	if err != nil {
		return orders, err
	}

	for index, key := range keys {
		orders[index].Id = key.IntID()
		products := make([]*Product, 0)
		err = json.Unmarshal(orders[index].ProductsSerial, &products)
		if err != nil {
			return nil, err
		}

		orders[index].Products = products
	}

	return orders, nil
}
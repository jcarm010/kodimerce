package entities

import (
	"time"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

const ENTITY_PRODUCT = "product"

type Product struct {
	Id int64 `datastore:"-" json:"id"`
	Name string `datastore:"name" json:"name"`
	Quantity int `datastore:"quantity" json:"quantity"`
	Active bool `datastore:"active" json:"active"`
	Created time.Time `datastore:"created" json:"created"`
}

func NewProduct(name string) *Product {
	return &Product{
		Name: name,
		Created: time.Now(),
	}
}

func CreateProduct(ctx context.Context, name string) (*Product, error) {
	product := NewProduct(name)

	key, err := datastore.Put(ctx, datastore.NewIncompleteKey(ctx, ENTITY_PRODUCT, nil), product)
	if err != nil {
		return nil, err
	}

	product.Id = key.IntID()
	return product, nil
}

func ListProducts(ctx context.Context) ([]*Product, error) {
	products := make([]*Product, 0)
	keys, err := datastore.NewQuery(ENTITY_PRODUCT).GetAll(ctx, &products)
	if err != nil {
		return nil, err
	}

	for index, key := range keys {
		products[index].Id = key.IntID()
	}

	return products, err
}
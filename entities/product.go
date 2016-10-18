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
	PriceCents int64 `datastore:"price_cents" json:"price_cents"`
	Pictures []string `datastore:"pictures,noindex" json:"pictures"`
	Description string `datastore:"description,noindex" json:"description"`
	Created time.Time `datastore:"created" json:"created"`
}

func NewProduct(name string) *Product {
	return &Product{
		Name: name,
		Created: time.Now(),
		Pictures: make([]string,0),
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
		var product = products[index];
		product.Id = key.IntID()
		if product.Pictures == nil {
			product.Pictures = make([]string, 0)
		}
	}

	return products, err
}

func UpdateProduct(ctx context.Context, product *Product) error {
	key := datastore.NewKey(ctx, ENTITY_PRODUCT, "", product.Id, nil)
	err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		p := &Product{}
		err := datastore.Get(ctx, key, p)
		if err != nil {
			return err
		}

		p.Name = product.Name
		p.PriceCents = product.PriceCents
		p.Quantity = product.Quantity
		p.Active = product.Active
		p.Pictures = product.Pictures
		p.Description = product.Description
		_, err = datastore.Put(ctx, key, p)
		return err
	}, nil)

	if err != nil {
		return err
	}

	return nil
}
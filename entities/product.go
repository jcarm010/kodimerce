package entities

import (
	"time"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"fmt"
	"encoding/json"
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
	//these fields are here to help building the UI
	PriceLabel string `datastore:"-" json:"price_label"`
	Thumbnail string `datastore:"-" json:"thumbnail"`
	Last bool `datastore:"-" json:"-"`
}

func (p *Product) SetMissingDefaults () {
	if p.Pictures == nil {
		p.Pictures = make([]string, 0)
	}

	p.Thumbnail = "/assets/images/stock.jpeg"
	if len(p.Pictures) > 0 {
		p.Thumbnail = p.Pictures[0]
	}

	p.PriceLabel = fmt.Sprintf("%.2f", float64(p.PriceCents)/100)
}

func (p *Product) OutOfStock() bool {
	return p.Quantity <= 0
}

func (p *Product) String() string {
	return p.Name
}

func (p *Product) PicturesJson() string {
	bts, _ := json.Marshal(p.Pictures)
	return string(bts)
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
		product.SetMissingDefaults()
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

func DecreaseProductInventory(ctx context.Context, productId int64, quantity int) error {
	key := datastore.NewKey(ctx, ENTITY_PRODUCT, "", productId, nil)
	err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		p := &Product{}
		err := datastore.Get(ctx, key, p)
		if err != nil {
			return err
		}

		p.Quantity = p.Quantity - quantity
		if p.Quantity < 0 {
			//todo: maybe send some alert?
			p.Quantity = 0
		}

		_, err = datastore.Put(ctx, key, p)
		return err
	}, nil)

	if err != nil {
		return err
	}

	return nil
}

func GetProduct(ctx context.Context, productId int64) (*Product, error) {
	key := datastore.NewKey(ctx, ENTITY_PRODUCT, "", productId, nil)
	product := &Product{}
	err := datastore.Get(ctx, key, product)
	if err != nil {
		return nil, err
	}

	product.SetMissingDefaults()
	product.Id = key.IntID()
	return product, nil
}

func GetProducts(ctx context.Context, productIds []int64) ([]*Product, error) {
	productKeys := make([]*datastore.Key, len(productIds))
	for index, productId := range productIds {
		key := datastore.NewKey(ctx, ENTITY_PRODUCT, "", productId, nil)
		productKeys[index] = key
	}

	products := make([]*Product, len(productIds))
	err := datastore.GetMulti(ctx, productKeys, products)
	if err != nil {
		return nil, err
	}

	for index, product := range products {
		product.SetMissingDefaults()
		product.Id = productKeys[index].IntID()
	}

	return products, nil
}

func GetProductsInCategories(ctx context.Context, categories []*Category) ([]*Product, error){
	log.Debugf(ctx, "Finding products in categories: %+v", categories)
	query := datastore.NewQuery(ENTITY_CATEGORY_PRODUCT)
	keyLookup := map[int64]bool{}
	keys := make([]*datastore.Key, 0)
	for _, category := range categories {
		categoryProducts := make([]*CategoryProduct, 0)
		_, err := query.Filter("category_id=", category.Id).GetAll(ctx, &categoryProducts)
		if err != nil {
			return nil, err
		}

		log.Debugf(ctx, "Category products: %+v", categoryProducts)
		for _, categoryProduct := range categoryProducts {
			productId := categoryProduct.ProductId
			if keyLookup[productId] {
				continue
			}

			keys = append(keys, datastore.NewKey(ctx, ENTITY_PRODUCT, "", categoryProduct.ProductId, nil))
			keyLookup[productId] = true
		}
	}

	log.Debugf(ctx, "Keys: %+v", len(keys))
	products := make([]*Product, len(keys))
	err := datastore.GetMulti(ctx, keys, products)
	if err != nil {
		return nil, err
	}

	p := make([]*Product, 0)
	for index, product := range products {
		product.Id = keys[index].IntID()
		if product.Active {
			p = append(p, product)
		}

		product.SetMissingDefaults()
	}

	return p, nil
}
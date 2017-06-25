package entities

import (
	"time"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"fmt"
	"encoding/json"
	"html/template"
)

const ENTITY_PRODUCT = "product"

type Product struct {
	//todo: add details that are not part of the description if html shows up in search engine.
	Id int64 `datastore:"-" json:"id"`
	Name string `datastore:"name" json:"name"`
	IsInfinite          bool `datastore:"is_infinite" json:"is_infinite"`
	Quantity            int `datastore:"quantity" json:"quantity"`
	NoShipping          bool `datastore:"no_shipping" json:"no_shipping"`
	NeedsDate           bool `datastore:"needs_date" json:"needs_date"`
	NeedsTime           bool `datastore:"needs_time" json:"needs_time"`
	NeedsPickupLocation bool `datastore:"needs_pickup_location" json:"needs_pickup_location"`
	AvailableTimes      []AvailableTime `datastore:"available_times" json:"available_times"`
	HasPricingOptions   bool `datastore:"has_pricing_options" json:"has_pricing_options"`
	PricingOptions      []PricingOption `datastore:"pricing_options" json:"pricing_options"`
	Active              bool `datastore:"active" json:"active"`
	PriceCents          int64 `datastore:"price_cents" json:"price_cents"`
	Pictures            []string `datastore:"pictures,noindex" json:"pictures"`
	Description         template.HTML `datastore:"description,noindex" json:"description"`
	Created             time.Time `datastore:"created" json:"created"`
	//these fields are here to help building the UI
	PriceLabel string `datastore:"-" json:"price_label"`
	Thumbnail string `datastore:"-" json:"thumbnail"`
	Last bool `datastore:"-" json:"-"`
}

type PricingOption struct {
	Label string `datastore:"label" json:"label"`
	PriceCents int64 `datastore:"price_cents" json:"price_cents"`
}

type ByCheapestPrice []PricingOption
func (a ByCheapestPrice) Len() int           { return len(a) }
func (a ByCheapestPrice) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCheapestPrice) Less(i, j int) bool { return a[i].PriceCents < a[j].PriceCents }

type AvailableTime struct {
	Hour int `datastore:"hour" json:"hour"`
	Minute int `datastore:"minute" json:"minute"`
}

func (t *AvailableTime) String () string {
	hour := fmt.Sprintf("%v", t.Hour)
	if t.Hour < 10 {
		hour = fmt.Sprintf("0%v", t.Hour)
	}

	minute := fmt.Sprintf("%v", t.Minute)
	if t.Minute < 10 {
		minute = fmt.Sprintf("0%v", t.Minute)
	}


	amHour := t.Hour % 12
	amHourLabel := fmt.Sprintf("%v", amHour)
	if amHour < 10 {
		amHourLabel = fmt.Sprintf("0%v", amHour)
	}

	var amlabel = "AM"
	if t.Hour > 12 {
		amlabel = "PM"
	}

	return hour + ":" + minute + " (" + amHourLabel + ":" + minute + " " + amlabel + ")"
}

type ByAvailableTime []AvailableTime
func (a ByAvailableTime) Len() int           { return len(a) }
func (a ByAvailableTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByAvailableTime) Less(i, j int) bool { return a[i].Hour * 60 + a[i].Minute < a[j].Hour * 60 + a[j].Minute }

func (p *Product) SetMissingDefaults () {
	if p.Pictures == nil {
		p.Pictures = make([]string, 0)
	}

	p.Thumbnail = "/assets/images/stock.jpeg"
	if len(p.Pictures) > 0 {
		p.Thumbnail = p.Pictures[0]
	}

	p.PriceLabel = p.GetPricingLabel()
	if p.AvailableTimes == nil {
		p.AvailableTimes = make([]AvailableTime, 0)
	}

	if p.PricingOptions == nil {
		p.PricingOptions = []PricingOption{}
	}
}

func (p *Product) GetPricingLabel() string {
	priceCents := p.GetPriceCents()
	return fmt.Sprintf("%.2f", float64(priceCents)/100)
}

func (p *Product) GetPriceCents() int64 {
	priceCents := p.PriceCents
	if p.HasPricingOptions && len(p.PricingOptions) > 0 {
		priceCents = p.PricingOptions[0].PriceCents
	}

	return priceCents
}

func (p *Product) OutOfStock() bool {
	return !p.IsInfinite && p.Quantity <= 0
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
		AvailableTimes: make([]AvailableTime, 0),
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
		p.IsInfinite = product.IsInfinite
		p.NoShipping = product.NoShipping
		p.NeedsDate = product.NeedsDate
		p.NeedsTime = product.NeedsTime
		p.AvailableTimes = product.AvailableTimes
		p.NeedsPickupLocation = product.NeedsPickupLocation
		p.HasPricingOptions = product.HasPricingOptions
		p.PricingOptions = product.PricingOptions
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

		if p.IsInfinite {
			//Do not need to update quantities since this product is infinite.
			return nil
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
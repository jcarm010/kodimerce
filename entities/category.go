package entities

import (
	"time"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"fmt"
)

const ENTITY_CATEGORY = "category"
const ENTITY_CATEGORY_PRODUCT = "category_product"

type Category struct {
	Id int64 `datastore:"-" json:"id"`
	Name string `datastore:"name" json:"name"`
	Description string `datastore:"description,noindex" json:"description"`
	Created time.Time `datastore:"created" json:"created"`
}

func (c *Category) String() string {
	return c.Name
}

func NewCategory(name string) *Category {
	return &Category{
		Name: name,
		Created: time.Now(),
	}
}

type CategoryProduct struct {
	CategoryId int64 `datastore:"category_id" json:"category_id"`
	ProductId int64 `datastore:"product_id" json:"product_id"`
}
func (cp *CategoryProduct) String() string {
	return fmt.Sprintf("%v-%v", cp.CategoryId, cp.ProductId)
}

func ListCategories(ctx context.Context) ([]*Category, error) {
	categories := make([]*Category, 0)
	keys, err := datastore.NewQuery(ENTITY_CATEGORY).GetAll(ctx, &categories)
	if err != nil {
		return nil, err
	}

	for index, key := range keys {
		var category = categories[index];
		category.Id = key.IntID()
	}

	return categories, err
}

func CreateCategory(ctx context.Context, name string) (*Category, error) {
	category := NewCategory(name)

	key, err := datastore.Put(ctx, datastore.NewIncompleteKey(ctx, ENTITY_CATEGORY, nil), category)
	if err != nil {
		return nil, err
	}

	category.Id = key.IntID()
	return category, nil
}

func UpdateCategory(ctx context.Context, category *Category) error {
	key := datastore.NewKey(ctx, ENTITY_CATEGORY, "", category.Id, nil)
	_, err := datastore.Put(ctx, key, category)
	if err != nil {
		return err
	}

	return nil
}

func SetCategoryProducts(ctx context.Context, categoryProducts []*CategoryProduct) error {
	keys := make([]*datastore.Key,0)
	for _, cp := range categoryProducts {
		keys = append(keys,datastore.NewKey(ctx, ENTITY_CATEGORY_PRODUCT, fmt.Sprintf("%v_%v", cp.CategoryId, cp.ProductId), 0, nil))
	}

	_, err := datastore.PutMulti(ctx, keys, categoryProducts)
	if err != nil {
		return err
	}

	return nil
}

func UnsetCategoryProducts(ctx context.Context, categoryProducts []*CategoryProduct) error {
	keys := make([]*datastore.Key,0)
	for _, cp := range categoryProducts {
		keys = append(keys,datastore.NewKey(ctx, ENTITY_CATEGORY_PRODUCT, fmt.Sprintf("%v_%v", cp.CategoryId, cp.ProductId), 0, nil))
	}

	err := datastore.DeleteMulti(ctx, keys)
	if err != nil {
		return err
	}

	return nil
}

func GetCategoryProducts(ctx context.Context) ([]*CategoryProduct, error) {
	categoryProducts := make([]*CategoryProduct, 0)
	_, err := datastore.NewQuery(ENTITY_CATEGORY_PRODUCT).GetAll(ctx, &categoryProducts)
	if err != nil {
		return nil, err
	}

	return categoryProducts, nil
}

func GetCategoryByName(ctx context.Context, name string) ([]*Category, error) {
	categories := make([]*Category, 0)
	query := datastore.NewQuery(ENTITY_CATEGORY)
	if name != "" {
		query = query.Filter("name=", name)
	}

	keys, err := query.GetAll(ctx, &categories)
	if err != nil {
		return nil, err
	}

	for index, category := range categories {
		category.Id = keys[index].IntID()
	}
	return categories, nil
}
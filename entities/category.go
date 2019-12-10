package entities

import (
	"fmt"
	"github.com/jcarm010/kodimerce/datastore"
	"golang.org/x/net/context"
	"strings"
	"time"
)

const EntityCategory = "category"
const EntityCategoryProduct = "category_product"

type Category struct {
	Id              int64     `datastore:"-" json:"id"`
	Name            string    `datastore:"name" json:"name"`
	Path            string    `datastore:"path" json:"path"`
	Description     string    `datastore:"description,noindex" json:"description"`
	MetaDescription string    `datastore:"meta_description,noindex" json:"meta_description"`
	Created         time.Time `datastore:"created" json:"created"`
	Thumbnail       string    `datastore:"thumbnail,noindex" json:"thumbnail"`
	Featured        bool      `datastore:"featured" json:"featured"`
}

func (c *Category) SetMissingDefaults() {
	if c.Thumbnail == "" {
		c.Thumbnail = "/assets/images/stock.jpeg"
	}

	if c.Path == "" {
		c.Path = c.Name
	}
}

func (c *Category) String() string {
	return c.Name
}

func NewCategory(name string) *Category {
	return &Category{
		Name:      name,
		Created:   time.Now(),
		Thumbnail: "/assets/images/stock.jpeg",
	}
}

type CategoryProduct struct {
	CategoryId int64 `datastore:"category_id" json:"category_id"`
	ProductId  int64 `datastore:"product_id" json:"product_id"`
}

func (cp *CategoryProduct) String() string {
	return fmt.Sprintf("%v-%v", cp.CategoryId, cp.ProductId)
}

func ListCategories(ctx context.Context) ([]*Category, error) {
	categories := make([]*Category, 0)
	keys, err := datastore.GetAll(ctx, datastore.NewQuery(EntityCategory), &categories)
	if err != nil {
		return nil, err
	}

	for index, key := range keys {
		var category = categories[index];
		category.Id = key.IntID()
		category.SetMissingDefaults()
	}

	return categories, err
}

func CreateCategory(ctx context.Context, name string) (*Category, error) {
	c := NewCategory(name)
	c.Path = name
	c.Path = strings.TrimSpace(c.Path)
	c.Path = strings.ToLower(c.Path)
	c.Path = strings.Replace(c.Path, " ", "-", -1)
	c.Path = strings.Replace(c.Path, "'", "", -1)

	key, err := datastore.Put(ctx, datastore.NewIncompleteKey(ctx, EntityCategory, nil), c)
	if err != nil {
		return nil, err
	}

	c.Id = key.IntID()
	return c, nil
}

func UpdateCategory(ctx context.Context, category *Category) error {
	key := datastore.NewKey(ctx, EntityCategory, "", category.Id, nil)
	_, err := datastore.Put(ctx, key, category)
	if err != nil {
		return err
	}

	return nil
}

func SetCategoryProducts(ctx context.Context, categoryProducts []*CategoryProduct) error {
	keys := make([]*datastore.Key, 0)
	for _, cp := range categoryProducts {
		keys = append(keys, datastore.NewKey(ctx, EntityCategoryProduct, fmt.Sprintf("%v_%v", cp.CategoryId, cp.ProductId), 0, nil))
	}

	_, err := datastore.PutMulti(ctx, keys, categoryProducts)
	if err != nil {
		return err
	}

	return nil
}

func UnsetCategoryProducts(ctx context.Context, categoryProducts []*CategoryProduct) error {
	keys := make([]*datastore.Key, 0)
	for _, cp := range categoryProducts {
		keys = append(keys, datastore.NewKey(ctx, EntityCategoryProduct, fmt.Sprintf("%v_%v", cp.CategoryId, cp.ProductId), 0, nil))
	}

	err := datastore.DeleteMulti(ctx, keys)
	if err != nil {
		return err
	}

	return nil
}

func GetCategoryProducts(ctx context.Context) ([]*CategoryProduct, error) {
	categoryProducts := make([]*CategoryProduct, 0)
	_, err := datastore.GetAll(ctx, datastore.NewQuery(EntityCategoryProduct), &categoryProducts)
	if err != nil {
		return nil, err
	}

	return categoryProducts, nil
}

func ListCategoriesByName(ctx context.Context, name string) ([]*Category, error) {
	categories := make([]*Category, 0)
	query := datastore.NewQuery(EntityCategory)
	if name != "" {
		query = query.Filter("name=", name)
	}

	keys, err := datastore.GetAll(ctx, query, &categories)
	if err != nil {
		return nil, err
	}

	for index, category := range categories {
		category.Id = keys[index].IntID()
		category.SetMissingDefaults()
	}

	return categories, nil
}

func ListCategoriesByPath(ctx context.Context, path string) ([]*Category, error) {
	categories := make([]*Category, 0)
	query := datastore.NewQuery(EntityCategory)
	if path != "" {
		query = query.Filter("path=", path)
	}

	keys, err := datastore.GetAll(ctx, query, &categories)
	if err != nil {
		return nil, err
	}

	for index, category := range categories {
		category.Id = keys[index].IntID()
		category.SetMissingDefaults()
	}

	return categories, nil
}

func ListCategoriesByFeatured(ctx context.Context, featured bool) ([]*Category, error) {
	categories := make([]*Category, 0)
	keys, err := datastore.GetAll(ctx, datastore.NewQuery(EntityCategory).Filter("featured=", featured), &categories)
	if err != nil {
		return nil, err
	}

	for index, key := range keys {
		var category = categories[index];
		category.Id = key.IntID()
		category.SetMissingDefaults()
	}

	return categories, err
}

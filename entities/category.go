package entities

import (
	"time"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

const ENTITY_CATEGORY = "category"

type Category struct {
	Id int64 `datastore:"-" json:"id"`
	Name string `datastore:"name" json:"name"`
	Description string `datastore:"description,noindex" json:"description"`
	Created time.Time `datastore:"created" json:"created"`
}

func NewCategory(name string) *Category {
	return &Category{
		Name: name,
		Created: time.Now(),
	}
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



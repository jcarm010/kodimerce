package entities

import (
	"time"
	"errors"
	"html/template"
	"strings"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"encoding/json"
	"fmt"
)

const ENTITY_GALLERY = "gallery"
var (
	ErrGalleryNotFound = errors.New("Not Found.")
)

type Gallery struct {
	Id int64 `datastore:"-" json:"id"`
	Title string `datastore:"title" json:"title"`
	Path string `datastore:"path" json:"path"`
	Description template.HTML `datastore:"description,noindex" json:"description"`
	MetaDescription string `datastore:"meta_description,noindex" json:"meta_description"`
	Published bool `datastore:"published" json:"published"`
	PublishedDate time.Time `datastore:"published_date" json:"published_date"`
	Images []*Image `datastore:"-" json:"images"`
	Created time.Time `datastore:"created" json:"created"`
	ImagesJson string `datastore:"images_json,noindex" json:"-"`
}

type Image struct {
	Url string `datastore:"url" json:"url"`
	AltTag string `datastore:"alt_tag" json:"alt_tag"`
}

func (p *Gallery) FormattedPublishedDate () (string) {
	return p.PublishedDate.Format("_2 Jan 2006")
}

func (p *Gallery) SetMissingDefaults () {
	images := make([]*Image, 0)
	json.Unmarshal([]byte(p.ImagesJson), &images)
	p.Images = images
}

func (p *Gallery) FirstImage() *Image {
	if len(p.Images) == 0 {
		return &Image{Url:"/assets/images/stock.jpeg", AltTag:"stock image"}
	}else {
		return p.Images[0]
	}
}

func NewGallery(title string) *Gallery {
	return &Gallery{
		Title: title,
		Created: time.Now(),
		Images: make([]*Image, 0),
	}
}

func CreateGallery(ctx context.Context, title string) (*Gallery, error) {
	p := NewGallery(title)
	p.Path = title
	p.Path = strings.TrimSpace(p.Path)
	p.Path = strings.ToLower(p.Path)
	p.Path = strings.Replace(p.Path, " ", "-", -1)
	p.Path = strings.Replace(p.Path, "'", "", -1)
	bts, err := json.Marshal(p.Images)
	if err != nil {
		return nil, fmt.Errorf("Error marshaling images: %s", err)
	}

	p.ImagesJson = fmt.Sprintf("%s", bts)
	key, err := datastore.Put(ctx, datastore.NewIncompleteKey(ctx, ENTITY_GALLERY, nil), p)
	if err != nil {
		return nil, err
	}

	p.SetMissingDefaults()
	p.Id = key.IntID()
	return p, nil
}

func ListGalleries(ctx context.Context, published bool, limit int) ([]*Gallery, error) {
	galleries := make([]*Gallery, 0)
	if limit == 0 {
		return galleries, nil
	}

	q := datastore.NewQuery(ENTITY_GALLERY)
	if published {
		q = q.Filter("published=", published)
	}

	if limit >= 0 {
		q = q.Limit(limit)
	}

	keys, err := q.GetAll(ctx, &galleries)
	if err != nil {
		return nil, err
	}

	for index, key := range keys {
		var gallery = galleries[index]
		gallery.Id = key.IntID()
		gallery.SetMissingDefaults()
	}

	return galleries, err
}

func UpdateGallery(ctx context.Context, gallery *Gallery) error {
	bts, err := json.Marshal(gallery.Images)
	if err != nil {
		return fmt.Errorf("Error marshaling images: %s", err)
	}

	gallery.ImagesJson = fmt.Sprintf("%s", bts)
	key := datastore.NewKey(ctx, ENTITY_GALLERY, "", gallery.Id, nil)
	err = datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		p := &Gallery{}
		err := datastore.Get(ctx, key, p)
		if err != nil {
			return err
		}

		p.Title = gallery.Title
		p.Path = gallery.Path
		p.Description = gallery.Description
		p.MetaDescription = gallery.MetaDescription
		if !p.Published && gallery.Published {
			p.PublishedDate = time.Now()
		}

		p.Published = gallery.Published
		p.ImagesJson = gallery.ImagesJson
		_, err = datastore.Put(ctx, key, p)
		return err
	}, nil)

	if err != nil {
		return err
	}

	return nil
}

func GetGalleryByPath(ctx context.Context, path string) (*Gallery, error) {
	galleries := make([]*Gallery, 0)
	keys, err := datastore.NewQuery(ENTITY_GALLERY).
		Filter("path=", path).
		Limit(1).
		GetAll(ctx, &galleries)

	if err != nil {
		return nil, err
	}

	if len(galleries) == 0 {
		return nil, ErrGalleryNotFound
	}

	key := keys[0]
	p := galleries[0]
	p.SetMissingDefaults()
	p.Id = key.IntID()
	return p, nil
}
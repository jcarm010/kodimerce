package entities

import (
	"time"
	"errors"
	"html/template"
	"strings"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

const (
	ENTITY_PAGE = "page"
	PROVIDER_SHALLOW_MIRROR = "shallow_mirror"
)
var (
	ErrPageNotFound = errors.New("Not Found.")
)

type Page struct {
	Id int64 `datastore:"-" json:"id"`
	Title string `datastore:"title" json:"title"`
	Provider string `datastore:"provider" json:"provider"`
	Path string `datastore:"path" json:"path"`
	Content template.HTML `datastore:"content,noindex" json:"content"`
	MetaDescription string `datastore:"meta_description,noindex" json:"meta_description"`
	Published bool `datastore:"published" json:"published"`
	PublishedDate time.Time `datastore:"published_date" json:"published_date"`
	Created time.Time `datastore:"created" json:"created"`
	ShallowMirrorUrl string `datastore:"shallow_mirror_url,noindex" json:"shallow_mirror_url"`
}

func (p *Page) FormattedPublishedDate () (string) {
	return p.PublishedDate.Format("_2 Jan 2006")
}

func (p *Page) SetMissingDefaults () {

}

func NewShallowMirrorPage(title string) *Page {
	return &Page{
		Title: title,
		Created: time.Now(),
		Provider: PROVIDER_SHALLOW_MIRROR,
	}
}

func CreatePage(ctx context.Context, title string) (*Page, error) {
	p := NewShallowMirrorPage(title)
	p.Path = title
	p.Path = strings.TrimSpace(p.Path)
	p.Path = strings.ToLower(p.Path)
	p.Path = strings.Replace(p.Path, " ", "-", -1)
	p.Path = strings.Replace(p.Path, "'", "", -1)
	key, err := datastore.Put(ctx, datastore.NewIncompleteKey(ctx, ENTITY_PAGE, nil), p)
	if err != nil {
		return nil, err
	}

	p.SetMissingDefaults()
	p.Id = key.IntID()
	return p, nil
}

func ListPages(ctx context.Context, published bool, limit int) ([]*Page, error) {
	pages := make([]*Page, 0)
	if limit == 0 {
		return pages, nil
	}

	q := datastore.NewQuery(ENTITY_PAGE)
	if published {
		q = q.Filter("published=", published)
	}

	if limit >= 0 {
		q = q.Limit(limit)
	}

	keys, err := q.GetAll(ctx, &pages)
	if err != nil {
		return nil, err
	}

	for index, key := range keys {
		var page = pages[index]
		page.Id = key.IntID()
		page.SetMissingDefaults()
	}

	return pages, err
}

func UpdatePage(ctx context.Context, page *Page) error {
	key := datastore.NewKey(ctx, ENTITY_PAGE, "", page.Id, nil)
	err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		p := &Page{}
		err := datastore.Get(ctx, key, p)
		if err != nil {
			return err
		}

		p.Title = page.Title
		p.Path = page.Path
		p.Content = page.Content
		p.MetaDescription = page.MetaDescription
		p.ShallowMirrorUrl = page.ShallowMirrorUrl
		if !p.Published && page.Published {
			p.PublishedDate = time.Now()
		}

		p.Published = page.Published
		_, err = datastore.Put(ctx, key, p)
		return err
	}, nil)

	if err != nil {
		return err
	}

	return nil
}


func GetPageByPath(ctx context.Context, path string) (*Page, error) {
	pages := make([]*Page, 0)
	keys, err := datastore.NewQuery(ENTITY_PAGE).
		Filter("path=", path).
		Limit(1).
		GetAll(ctx, &pages)

	if err != nil {
		return nil, err
	}

	if len(pages) == 0 {
		return nil, ErrPageNotFound
	}

	key := keys[0]
	p := pages[0]
	p.SetMissingDefaults()
	p.Id = key.IntID()
	return p, nil
}
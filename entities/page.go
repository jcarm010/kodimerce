package entities

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jcarm010/kodimerce/datastore"
	"golang.org/x/net/context"
	"html/template"
	"strings"
	"time"
)

const (
	EntityPage            = "page"
	ProviderShallowMirror = "shallow_mirror"
	ProviderCustomPage    = "dynamic_page"
	ProviderRedirectPage  = "redirect_page"
)

var (
	ErrPageNotFound = errors.New("not found")
)

type Page struct {
	Id                 int64         `datastore:"-" json:"id"`
	Title              string        `datastore:"title" json:"title"`
	Provider           string        `datastore:"provider" json:"provider"`
	Path               string        `datastore:"path" json:"path"`
	Content            template.HTML `datastore:"content,noindex" json:"content"`
	MetaDescription    string        `datastore:"meta_description,noindex" json:"meta_description"`
	Published          bool          `datastore:"published" json:"published"`
	PublishedDate      time.Time     `datastore:"published_date" json:"published_date"`
	Created            time.Time     `datastore:"created" json:"created"`
	ShallowMirrorUrl   string        `datastore:"shallow_mirror_url,noindex" json:"shallow_mirror_url"`
	DynamicPage        *DynamicPage  `datastore:"-" json:"dynamic_page"`
	RawDynamicPage     []byte        `datastore:"raw_dynamic_page,noindex" json:"-"`
	RedirectUrl        string        `datastore:"redirect_url,noindex" json:"redirect_url"`
	RedirectStatusCode int           `datastore:"redirect_status_code,noindex" json:"redirect_status_code"`
}

type DynamicPage struct {
	Title           string                     `json:"title"`
	MetaDescription string                     `json:"meta_description"`
	HasNavigation   bool                       `json:"has_navigation"`
	HasBanner       bool                       `json:"has_banner"`
	Banner          *DynamicPageImageComponent `json:"banner"`
	Rows            []*DynamicPageRow          `json:"rows"`
}

type DynamicPageRow struct {
	ComponentName      string                         `json:"component_name"` //this should be the name of the component to use
	RowSimpleComponent *DynamicPageRowSimpleComponent `json:"row_simple_component"`
	SeparatorTop       bool                           `json:"separator_top"`
	SeparatorBottom    bool                           `json:"separator_bottom"`
}

type DynamicPageImageComponent struct {
	Path    string `json:"path"`
	AltText string `json:"alt_text"`
	SetSize bool   `json:"set_size"`
	Width   string `json:"width"`
	Height  string `json:"height"`
}

type DynamicPageRowSimpleComponent struct {
	Header        string                     `json:"header"`
	IsMainHeader  bool                       `json:"is_main_header"`
	Description   template.HTML              `json:"description"`
	HasImage      bool                       `json:"has_image"`
	Image         *DynamicPageImageComponent `json:"image"`
	ImagePosition string                     `json:"image_position"`
}

func NewDynamicPage(title string, metaDescription string) *DynamicPage {
	return &DynamicPage{
		Title:           title,
		MetaDescription: metaDescription,
		Rows:            make([]*DynamicPageRow, 0),
	}
}

func (p *Page) String() string {
	bts, _ := json.Marshal(p)
	return fmt.Sprintf("%s", bts)
}

func (p *Page) FormattedPublishedDate() (string) {
	return p.PublishedDate.Format("_2 Jan 2006")
}

func (p *Page) SetMissingDefaults() {
	if p.RawDynamicPage != nil && len(p.RawDynamicPage) != 0 {
		dynamicPage := &DynamicPage{}
		fmt.Printf("Marshalling dynamic page: %s\n", p.RawDynamicPage)
		err := json.Unmarshal(p.RawDynamicPage, dynamicPage)
		if err == nil {
			p.DynamicPage = dynamicPage
		}

		for _, row := range dynamicPage.Rows {
			if row.RowSimpleComponent.ImagePosition == "" {
				row.RowSimpleComponent.ImagePosition = "top"
			}
		}
	}

	if p.DynamicPage == nil {
		p.DynamicPage = NewDynamicPage(p.Title, "")
	}
}

func NewShallowMirrorPage(title string) *Page {
	return &Page{
		Title:       title,
		Created:     time.Now(),
		Provider:    ProviderShallowMirror,
		DynamicPage: NewDynamicPage(title, ""),
	}
}

func NewPage(title string, metaDescription string) *Page {
	return &Page{
		Title:       title,
		Created:     time.Now(),
		Provider:    ProviderCustomPage,
		DynamicPage: NewDynamicPage(title, metaDescription),
	}
}

func CreatePage(ctx context.Context, title string) (*Page, error) {
	p := NewShallowMirrorPage(title)
	p.Path = title
	p.Path = strings.TrimSpace(p.Path)
	p.Path = strings.ToLower(p.Path)
	p.Path = strings.Replace(p.Path, " ", "-", -1)
	p.Path = strings.Replace(p.Path, "'", "", -1)
	key, err := datastore.Put(ctx, datastore.NewIncompleteKey(ctx, EntityPage, nil), p)
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

	q := datastore.NewQuery(EntityPage)
	if published {
		q = q.Filter("published=", published)
	}

	if limit >= 0 {
		q = q.Limit(limit)
	}

	keys, err := datastore.GetAll(ctx, q, &pages)
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
	key := datastore.NewKey(ctx, EntityPage, "", page.Id, nil)
	err := datastore.RunInTransaction(ctx, func(transaction *datastore.Transaction) error {
		p := &Page{}
		err := transaction.Get(key, p)
		if err != nil {
			return err
		}

		p.Provider = page.Provider
		p.Title = page.Title
		p.Path = page.Path
		p.Content = page.Content
		p.MetaDescription = page.MetaDescription
		p.ShallowMirrorUrl = page.ShallowMirrorUrl
		p.RedirectStatusCode = page.RedirectStatusCode
		p.RedirectUrl = page.RedirectUrl
		if !p.Published && page.Published {
			p.PublishedDate = time.Now()
		}

		bts, err := json.Marshal(page.DynamicPage)
		if err != nil {
			return err
		}

		p.RawDynamicPage = bts
		p.Published = page.Published
		_, err = transaction.Put(key, p)
		return err
	})

	if err != nil {
		return err
	}

	return nil
}

func GetPageByPath(ctx context.Context, path string) (*Page, error) {
	pages := make([]*Page, 0)
	keys, err := datastore.GetAll(ctx, datastore.NewQuery(EntityPage).
		Filter("path=", path).
		Limit(1), &pages)

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

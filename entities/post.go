package entities

import (
	"time"
	"html/template"
	"strings"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"errors"
)

const ENTITY_POST = "post"
var (
	ErrPostNotFound = errors.New("Not Found.")
)

type Post struct {
	Id int64 `datastore:"-" json:"id"`
	Title string `datastore:"title" json:"title"`
	Path string `datastore:"path" json:"path"`
	Content template.HTML `datastore:"content,noindex" json:"content"`
	ShortDescription string `datastore:"short_description,noindex" json:"short_description"`
	MetaDescription string `datastore:"meta_description,noindex" json:"meta_description"`
	Banner string `datastore:"banner,noindex" json:"banner"`
	Published bool `datastore:"published" json:"published"`
	PublishedDate time.Time `datastore:"published_date" json:"published_date"`
	UpdatedDate time.Time `datastore:"updated_date" json:"updated_date"`
	Created time.Time `datastore:"created" json:"created"`
}

type ByNewestFirst []*Post
func (a ByNewestFirst) Len() int           { return len(a) }
func (a ByNewestFirst) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByNewestFirst) Less(i, j int) bool { return a[j].PublishedDate.Unix() < a[i].PublishedDate.Unix() }

func (p *Post) FormattedPublishedDate () (string) {
	return p.FormattedDateDMY(p.PublishedDate)
}

func (p *Post) FormattedDateDMY (dte time.Time) (string) {
	return dte.Format("_2 Jan 2006")
}

func (p *Post) SetMissingDefaults () {
	t := time.Time{}
	if p.UpdatedDate == t {
		p.UpdatedDate = p.PublishedDate
	}
}

func NewPost(title string) *Post {
	return &Post{
		Title: title,
		Created: time.Now(),
	}
}

func CreatePost(ctx context.Context, title string) (*Post, error) {
	p := NewPost(title)
	p.Path = title
	p.Path = strings.TrimSpace(p.Path)
	p.Path = strings.ToLower(p.Path)
	p.Path = strings.Replace(p.Path, " ", "-", -1)
	p.Path = strings.Replace(p.Path, "'", "", -1)
	key, err := datastore.Put(ctx, datastore.NewIncompleteKey(ctx, ENTITY_POST, nil), p)
	if err != nil {
		return nil, err
	}

	p.SetMissingDefaults()
	p.Id = key.IntID()
	return p, nil
}

func UpdatePost(ctx context.Context, post *Post) error {
	key := datastore.NewKey(ctx, ENTITY_POST, "", post.Id, nil)
	err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		p := &Post{}
		err := datastore.Get(ctx, key, p)
		if err != nil {
			return err
		}

		p.Title = post.Title
		p.Path = post.Path
		p.Content = post.Content
		p.MetaDescription = post.MetaDescription
		p.Banner = post.Banner
		p.ShortDescription = post.ShortDescription
		p.PublishedDate = post.PublishedDate
		if !p.Published && post.Published {
			p.PublishedDate = time.Now()
		}
		p.UpdatedDate = time.Now()
		p.Published = post.Published
		_, err = datastore.Put(ctx, key, p)
		return err
	}, nil)

	if err != nil {
		return err
	}

	return nil
}

func ListPosts(ctx context.Context, published bool, limit int) ([]*Post, error) {
	posts := make([]*Post, 0)
	if limit == 0 {
		return posts, nil
	}

	q := datastore.NewQuery(ENTITY_POST)
	if published {
		q = q.Filter("published=", published).
			Order("-published_date")
	}else {
		q = q.Order("-created")
	}

	if limit >= 0 {
		q = q.Limit(limit)
	}

	keys, err := q.GetAll(ctx, &posts)
	if err != nil {
		return nil, err
	}

	for index, key := range keys {
		var post = posts[index]
		post.Id = key.IntID()
		post.SetMissingDefaults()
	}

	return posts, err
}

func GetPostByPath(ctx context.Context, path string) (*Post, error) {
	posts := make([]*Post, 0)
	keys, err := datastore.NewQuery(ENTITY_POST).
		Filter("path=", path).
		Limit(1).
		GetAll(ctx, &posts)

	if err != nil {
		return nil, err
	}
	if len(posts) == 0 {
		return nil, ErrPostNotFound
	}

	key := keys[0]
	p := posts[0]
	p.SetMissingDefaults()
	p.Id = key.IntID()
	return p, nil
}
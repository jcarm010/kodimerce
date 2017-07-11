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
	MetaDescription string `datastore:"meta_description,noindex" json:"meta_description"`
	Banner string `datastore:"banner,noindex" json:"banner"`
	Published bool `datastore:"published" json:"published"`
	PublishedDate time.Time `datastore:"published_date" json:"published_date"`
	Created time.Time `datastore:"created" json:"created"`
}

func (p *Post) FormattedPublishedDate () (string) {
	return p.PublishedDate.Format("_2 Jan 2006")
}

func (p *Post) SetMissingDefaults () {
	if p.MetaDescription == "" {
		p.MetaDescription = string(p.Content)
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
		if !p.Published && post.Published {
			p.PublishedDate = time.Now()
		}

		p.Published = post.Published
		_, err = datastore.Put(ctx, key, p)
		return err
	}, nil)

	if err != nil {
		return err
	}

	return nil
}

func ListPosts(ctx context.Context) ([]*Post, error) {
	posts := make([]*Post, 0)
	q := datastore.NewQuery(ENTITY_POST)
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

func ListPostsByPublished(ctx context.Context, published bool) ([]*Post, error) {
	posts := make([]*Post, 0)
	q := datastore.NewQuery(ENTITY_POST).
		Filter("published=", published)
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
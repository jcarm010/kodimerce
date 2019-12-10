package search_api

import (
	"context"
	"google.golang.org/appengine/search"
	"strconv"
	"strings"
	"time"
)

const (
	BlobIndexName = "blobs"
)

type Client struct {
	Context context.Context
}

type SearchBlob struct {
	BlobKey      string    `datastore:"blob_key"`
	ContentType  string    `datastore:"content_type"`
	CreationTime time.Time `datastore:"creation"`
	Filename     string    `datastore:"filename"`
	MD5          string    `datastore:"md5_hash"`
	ObjectName   string    `datastore:"gs_object_name"`
	Size         string    `datastore:"size"`
	Title        string    `datastore:"title"`
}

func NewClient(ctx context.Context) Client {
	return Client{
		Context: ctx,
	}
}

func (s Client) Init(blob SearchBlob) error {
	index, err := search.Open(BlobIndexName)
	if err != nil {
		return err
	}

	blob.Title = createTitle(blob.Filename)
	_, err = index.Put(s.Context, blob.BlobKey, &blob)
	if err != nil {
		return err
	}

	return nil
}

func (s Client) PutBlob(blob SearchBlob) error {
	index, err := search.Open(BlobIndexName)
	if err != nil {
		return err
	}

	blob.Title = createTitle(blob.Filename)
	_, err = index.Put(s.Context, blob.BlobKey, &blob)
	if err != nil {
		return err
	}

	return nil
}

func (s Client) GetBlobs(searchKey string, limit int, cursorStr string) ([]*BlobInfo, string, int, error) {
	index, err := search.Open(BlobIndexName)
	if err != nil {
		return nil, "", 0, err
	}

	blobs := make([]*BlobInfo, 0)
	options := search.SearchOptions{
		Limit:  limit,
		Cursor: search.Cursor(cursorStr),
	}

	var cursor string
	var total int
	for t := index.Search(s.Context, searchKey, &options); ; {
		var temp SearchBlob
		id, err := t.Next(&temp)
		if err == search.Done {
			break
		}

		if err != nil {
			return nil, "", 0, err
		}

		sizeInt, err := strconv.Atoi(temp.Size)
		if err != nil {
			return nil, "", 0, err
		}

		blob := BlobInfo{
			BlobKey:      id,
			ContentType:  temp.ContentType,
			CreationTime: temp.CreationTime,
			Filename:     temp.Filename,
			Size:         int64(sizeInt),
			MD5:          temp.MD5,
		}

		blobs = append(blobs, &blob)
		cursor = string(t.Cursor())
		total = t.Count()
	}

	return blobs, cursor, total, nil

}

func (s Client) DeleteIndex(key string) error {
	index, err := search.Open(BlobIndexName)
	if err != nil {
		return err
	}

	err = index.Delete(s.Context, key)
	if err != nil {
		return err
	}

	return nil
}

func createTitle(fileName string) string {
	return strings.Replace(fileName, "-", " ", -1)
}

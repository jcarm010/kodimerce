package entities

import (
	"github.com/jcarm010/kodimerce/search_api"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/blobstore"
	"google.golang.org/appengine/datastore"
	"strconv"
	"strings"
)

type BlobResponse struct {
	Blobs  []*blobstore.BlobInfo `json:"blobs"`
	Cursor string                `json:"cursor"`
	Total  int                   `json:"total"`
}

const ENTITY_BLOB = "__BlobInfo__"

func InitSearchAPI(ctx context.Context) error {
	blobs := make([]*blobstore.BlobInfo, 0)
	keys, err := datastore.NewQuery(ENTITY_BLOB).GetAll(ctx, &blobs)
	if err != nil {
		index := strings.Index(err.Error(), "datastore: cannot load field")
		if index != 0 {
			return err
		}
	}

	searchClient := search_api.NewClient(ctx)
	for i, blob := range blobs {
		sizeString := strconv.Itoa(int(blob.Size))
		searchBlob := search_api.SearchBlob{
			BlobKey:      keys[i].StringID(),
			ContentType:  blob.ContentType,
			CreationTime: blob.CreationTime,
			Filename:     blob.Filename,
			MD5:          blob.MD5,
			ObjectName:   blob.ObjectName,
			Size:         sizeString,
		}

		err = searchClient.Init(searchBlob)
		if err != nil {
			return err
		}
	}

	return nil
}

func ListUploads(ctx context.Context, cursorStr string, limit int, search string) (*BlobResponse, error) {
	blobs := make([]*blobstore.BlobInfo, 0)
	var err error
	var total int
	if search != "" {
		searchClient := search_api.NewClient(ctx)
		blobs, cursorStr, total, err = searchClient.GetBlobs(search, limit, cursorStr)
		if err != nil {
			return nil, err
		}

	} else {
		total, err = datastore.NewQuery(ENTITY_BLOB).Count(ctx)
		if err != nil {
			return nil, err
		}

		query := datastore.NewQuery(ENTITY_BLOB).Limit(limit)
		if cursorStr != "" {
			cursor, err := datastore.DecodeCursor(cursorStr)
			if err != nil {
				return nil, err
			}

			query = query.Start(cursor)
		}

		t := query.Run(ctx)
		for {
			var blob blobstore.BlobInfo
			key, err := t.Next(&blob)
			if err == datastore.Done {
				break
			}

			if err != nil {
				return nil, err
			}

			blob.BlobKey = appengine.BlobKey(key.StringID())
			blobs = append(blobs, &blob)
		}

		if cursor, err := t.Cursor(); err == nil {
			cursorStr = cursor.String()
			if err != nil {
				return nil, err
			}
		}
	}

	blobResp := BlobResponse{
		Blobs:  blobs,
		Cursor: cursorStr,
		Total:  total,
	}

	return &blobResp, nil
}

func GetUpload(ctx context.Context, key appengine.BlobKey) (*blobstore.BlobInfo, error) {
	blob := &blobstore.BlobInfo{}
	k := datastore.NewKey(ctx, ENTITY_BLOB, string(key), 0, nil)
	err := datastore.Get(ctx, k, blob)
	if err != nil {
		index := strings.Index(err.Error(), "datastore: cannot load field")
		if index != 0 {
			return nil, err
		}
	}

	blob.BlobKey = key
	return blob, nil
}

func GetUploadByName(ctx context.Context, name string) (*blobstore.BlobInfo, error) {
	blobs := make([]*blobstore.BlobInfo, 0)
	keys, err := datastore.NewQuery(ENTITY_BLOB).Filter("filename=", name).Limit(1).GetAll(ctx, &blobs)
	if err != nil {
		index := strings.Index(err.Error(), "datastore: cannot load field")
		if index != 0 {
			return nil, err
		}
	}

	if len(blobs) == 0 {
		return nil, datastore.ErrNoSuchEntity
	}

	for i, blob := range blobs {
		blob.BlobKey = appengine.BlobKey(keys[i].StringID())
	}

	return blobs[0], nil
}

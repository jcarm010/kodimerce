package entities

import (
	"github.com/jcarm010/kodimerce/datastore"
	"github.com/jcarm010/kodimerce/search_api"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
	"strconv"
	"strings"
)

type BlobResponse struct {
	Blobs  []*search_api.BlobInfo `json:"blobs"`
	Cursor string                 `json:"cursor"`
	Total  int                    `json:"total"`
}

const EntityBlob = "file_uploads"

func InitSearchAPI(ctx context.Context) error {
	blobs := make([]*search_api.BlobInfo, 0)
	keys, err := datastore.GetAll(ctx, datastore.NewQuery(EntityBlob), &blobs)
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
	search = "" //todo: search needs to be implemented without AppEngine search
	blobs := make([]*search_api.BlobInfo, 0)
	var err error
	var total int
	if search != "" {
		searchClient := search_api.NewClient(ctx)
		blobs, cursorStr, total, err = searchClient.GetBlobs(search, limit, cursorStr)
		if err != nil {
			return nil, err
		}

	} else {
		total, err = datastore.Count(ctx, datastore.NewQuery(EntityBlob))
		if err != nil {
			return nil, err
		}

		query := datastore.NewQuery(EntityBlob).Limit(limit)
		if cursorStr != "" {
			cursor, err := datastore.DecodeCursor(cursorStr)

			if err != nil {
				return nil, err
			}

			query = query.Start(cursor)
		}

		t := datastore.Run(ctx, query)
		for {
			var blob search_api.BlobInfo
			key, err := t.Next(&blob)
			if err == iterator.Done {
				break
			}

			if err != nil {
				return nil, err
			}

			blob.BlobKey = key.Name
			blobs = append(blobs, &blob)
		}

		if cursor, err := t.Cursor(); err == nil {
			cursorStr = cursor.String()
		}
	}

	blobResp := BlobResponse{
		Blobs:  blobs,
		Cursor: cursorStr,
		Total:  total,
	}

	return &blobResp, nil
}

func GetUpload(ctx context.Context, key string) (*search_api.BlobInfo, error) {
	blob := &search_api.BlobInfo{}
	k := datastore.NewKey(ctx, EntityBlob, key, 0, nil)
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

func GetUploadByName(ctx context.Context, name string) (*search_api.BlobInfo, error) {
	blobs := make([]*search_api.BlobInfo, 0)
	keys, err := datastore.GetAll(ctx, datastore.NewQuery(EntityBlob).Filter("filename=", name).Limit(1), &blobs)
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
		blob.BlobKey = keys[i].StringID()
	}

	return blobs[0], nil
}

func PutUpload(ctx context.Context, info *search_api.BlobInfo) error {
	key := datastore.NewKey(ctx, EntityBlob, info.BlobKey, 0, nil)
	_, err := datastore.Put(ctx, key, info)
	return err
}
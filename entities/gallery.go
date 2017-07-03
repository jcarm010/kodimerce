package entities

import (
	"google.golang.org/appengine/blobstore"
	"google.golang.org/appengine/datastore"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"strings"
)

const ENTITY_BLOB = "__BlobInfo__"

func ListUploads(ctx context.Context) ([]*blobstore.BlobInfo, error) {
	blobs := make([]*blobstore.BlobInfo, 0)
	keys, err := datastore.NewQuery(ENTITY_BLOB).GetAll(ctx, &blobs)
	if err != nil {
		index := strings.Index(err.Error(), "datastore: cannot load field")
		if index != 0 {
			return nil, err
		}
	}

	for i, blob := range blobs {
		blob.BlobKey = appengine.BlobKey(keys[i].StringID())
	}

	return blobs, nil
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
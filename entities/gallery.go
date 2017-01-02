package entities

import (
	"google.golang.org/appengine/blobstore"
	"google.golang.org/appengine/datastore"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
)

const ENTITY_BLOB = "__BlobInfo__"

func ListUploads(ctx context.Context) ([]*blobstore.BlobInfo, error) {
	blobs := make([]*blobstore.BlobInfo, 0)
	keys, err := datastore.NewQuery(ENTITY_BLOB).GetAll(ctx, &blobs)
	if err != nil {
		return nil, err
	}

	for i, blob := range blobs {
		blob.BlobKey = appengine.BlobKey(keys[i].StringID())
	}

	return blobs, err
}

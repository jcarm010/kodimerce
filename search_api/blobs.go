package search_api

import (
	"time"
)

type BlobInfo struct {
	BlobKey      string    `datastore:"-"`
	OldBlobKey   string    `datastore:"old_blob_key"`
	NewBlobKey   string    `datastore:"new_blob_key"`
	ContentType  string    `datastore:"content_type"`
	CreationTime time.Time `datastore:"creation"`
	Filename     string    `datastore:"filename"`
	Size         int64     `datastore:"size"`
	MD5          string    `datastore:"md5_hash"`
	UploadId     string    `datastore:"upload_id,omitempty"`

	// ObjectName is the Google Cloud Storage name for this blob.
	ObjectName string `datastore:"gs_object_name"`
}

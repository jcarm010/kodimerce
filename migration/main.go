package main

import (
	"context"
	"github.com/jcarm010/kodimerce/datastore"
	"github.com/jcarm010/kodimerce/log"
	"github.com/jcarm010/kodimerce/search_api"
)

type BlobKeyMapping struct {
	FileName   string `datastore:"gcs_filename"`
	NewBlobKey string `datastore:"new_blob_key"`
	OldBlobKey string `datastore:"old_blob_key"`
}

func main() {
	ctx := context.Background()
	blobKeyMappingDocuments := make([]BlobKeyMapping, 0)
	_, err := datastore.GetAll(ctx, datastore.NewQuery("_blobmigrator_BlobKeyMapping"), &blobKeyMappingDocuments)
	if err != nil {
		panic(err)
	}

	oldKeyMappings := make(map[string]BlobKeyMapping)
	for _, key := range blobKeyMappingDocuments {
		log.Infof(ctx, "Key=======================")
		log.Infof(ctx, "FileName: %s", key.FileName)
		log.Infof(ctx, "OldBlobKey: %s", key.OldBlobKey)
		log.Infof(ctx, "NewBlobKey: %s", key.NewBlobKey)
		oldKeyMappings[key.OldBlobKey] = key
	}

	blobInfos := make([]search_api.BlobInfo, 0)
	blobInfoKeys, err := datastore.GetAll(ctx, datastore.NewQuery("__BlobInfo__"), &blobInfos)
	if err != nil {
		panic(err)
	}

	newBlobKeys := make([]*datastore.Key, 0)
	newBlobInfos := make([]search_api.BlobInfo, 0)
	for index, blobInfo := range blobInfos {
		blobKey := blobInfoKeys[index].StringID()
		oldKeyMapping, exists := oldKeyMappings[blobKey]
		if !exists {
			log.Infof(ctx, "KeyMapping NOT Found for key: %s", blobKey)
			continue
		}

		newBlobInfos = append(newBlobInfos, search_api.BlobInfo{
			OldBlobKey: oldKeyMapping.OldBlobKey,
			NewBlobKey: oldKeyMapping.NewBlobKey,
			ContentType: blobInfo.ContentType,
			CreationTime: blobInfo.CreationTime,
			Filename: blobInfo.Filename,
			Size: blobInfo.Size,
			MD5: blobInfo.MD5,
			UploadId: blobInfo.UploadId,
			ObjectName: oldKeyMapping.FileName,
		})

		newBlobKeys = append(newBlobKeys, datastore.NewKey(ctx, "file_uploads", oldKeyMapping.OldBlobKey, 0, nil))
	}

	_, err = datastore.PutMulti(ctx, newBlobKeys, newBlobInfos)
	if err != nil {
		panic(err)
	}
}

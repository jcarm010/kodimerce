package storage

import (
	"cloud.google.com/go/storage"
	"context"
	"io"
	"os"
	"strings"
)
var (
	storageClient *storage.Client
	bucketName = os.Getenv("GOOGLE_CLOUD_PROJECT") + ".appspot.com"
)

func init() {
	var err error
	storageClient, err = storage.NewClient(context.Background())
	if err != nil {
		panic(err)
	}
}

func GetObject(objectName string) *storage.ObjectHandle {
	bucketPrefix := "/" + bucketName + "/"
	objectName = strings.TrimPrefix(objectName, bucketPrefix)
	return storageClient.Bucket(bucketName).Object(objectName)
}

func PutObject(ctx context.Context, objectName string, reader io.Reader) error {
	wc := storageClient.Bucket(bucketName).Object(objectName).NewWriter(ctx)
	if _, err := io.Copy(wc, reader); err != nil {
		return err
	}

	if err := wc.Close(); err != nil {
		return err
	}

	return nil
}
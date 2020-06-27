package storageclient

import (
	"errors"
	"io/ioutil"

	"cloud.google.com/go/storage"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

var ctx context.Context
var googleCloudStorageClient *storage.Client

// Meta fetchが返却する構造体
type Meta struct {
	Data        []byte
	Size        int64
	ContentType string
}

var (
	// ErrBucketNotFound bucketが見つからなかった
	ErrBucketNotFound = errors.New("bucket not found")

	// ErrBlobNotFound blobが見つからなかった
	ErrBlobNotFound = errors.New("blob not found")
)

func init() {
	sugar := zap.NewExample().Sugar()
	defer sugar.Sync()

	ctx = context.Background() // グローバルのctxに代入
	client, err := storage.NewClient(ctx)
	if err != nil {
		sugar.Fatalw("Failed to create client", "error", err)
	}
	googleCloudStorageClient = client
}

// FetchBlob bucketNameのbucketからblobNameのblobを取ってきてMetaの形で返す
func FetchBlob(bucketName string, blobName string) (*Meta, error) {
	bucket := googleCloudStorageClient.Bucket(bucketName)

	blob := bucket.Object(blobName)
	reader, err := blob.NewReader(ctx)
	if err != nil {
		switch err {
		case storage.ErrBucketNotExist:
			return nil, ErrBucketNotFound
		case storage.ErrObjectNotExist:
			return nil, ErrBlobNotFound
		default:
			return nil, err
		}
	}
	defer reader.Close()

	blobAttrs, err := blob.Attrs(ctx)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	meta := Meta{
		Data:        data,
		Size:        blobAttrs.Size,
		ContentType: blobAttrs.ContentType,
	}
	return &meta, nil
}

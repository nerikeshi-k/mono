package provider

import (
	"bytes"
	"errors"
	"io"
	"mime"
	"os"
	"path"
	"strings"
	"time"

	"github.com/nerikeshi-k/mono/config"
	"github.com/nerikeshi-k/mono/env"
	"github.com/nerikeshi-k/mono/preprocess"
	"github.com/nerikeshi-k/mono/recordstore"
	"github.com/nerikeshi-k/mono/storageclient"
	"github.com/nerikeshi-k/mono/util"
	"golang.org/x/exp/slices"

	"go.uber.org/zap"
)

var (
	// ErrNotFound blobやbucketが見つからなかった
	ErrNotFound = errors.New("not found")

	// ErrInternalServerError 処理中の予期せぬエラー
	ErrInternalServerError = errors.New("internal server error")
)

// Product provideが返すもの
type Product struct {
	Data   []byte
	Record *recordstore.Record
}

func fetchRecord(bucketName string, blobName string) (*recordstore.Record, error) {
	sugar := zap.NewExample().Sugar()
	defer sugar.Sync()

	key := recordstore.GenerateKey(bucketName, blobName)
	record, err := recordstore.GetRecord(key)
	if err == nil && util.DoesFileExist(record.GetPath()) {
		if env.DEBUG {
			sugar.Debugw("cache hit")
		}
		record.LastRequestedAt = time.Now()
		recordstore.SetRecord(key, record)
		return record, nil
	}
	// gcsからblocを取ってくる
	blob, err := storageclient.FetchBlob(bucketName, blobName)
	if err != nil {
		sugar.Errorw("failed to fetch blob", "error", err)
		return nil, err
	}
	// キャッシュ保存
	now := time.Now()
	cacheFileName := recordstore.GenerateCacheFileName()
	fp, err := os.Create(path.Join(config.Get().CacheDirPath, cacheFileName))
	if err != nil {
		sugar.Errorw("failed to create cache file", "error", err)
		return nil, err
	}
	defer fp.Close()
	if _, err := io.Copy(fp, bytes.NewReader(blob.Data)); err != nil {
		sugar.Errorw("failed to write cache file", "error", err)
		return nil, err
	}

	// Record作成、保存
	mediatype, _, err := mime.ParseMediaType(blob.ContentType)
	if err != nil {
		sugar.Errorw("failed to parse content type", "error", err)
		return nil, err
	}
	newRecord := &recordstore.Record{
		BlobName:        blobName,
		CacheFileName:   cacheFileName,
		Size:            blob.Size,
		ContentType:     mediatype,
		LastRequestedAt: now,
		CreatedAt:       now,
	}
	recordstore.SetRecord(key, newRecord)
	return newRecord, nil
}

func predictContentType(blobName string) (string, error) {
	if strings.HasSuffix(blobName, ".png") {
		return "image/png", nil
	} else if strings.HasSuffix(blobName, ".jpeg") || strings.HasSuffix(blobName, ".jpg") {
		return "image/jpeg", nil
	} else if strings.HasSuffix(blobName, ".webp") {
		return "image/webp", nil
	} else {
		return "", ErrInternalServerError
	}
}

// Provide bucketからblobを取ってきてProductにして返す
func Provide(bucketName string, blobName string, query preprocess.Query) (*Product, error) {
	sugar := zap.NewExample().Sugar()
	defer sugar.Sync()

	// bucket名がconfig内に指定されているか確認
	bucketNameExists := false
	for _, bucket := range config.Get().Buckets {
		if bucket.Name == bucketName {
			bucketNameExists = true
			break
		}
	}
	if !bucketNameExists {
		return nil, ErrNotFound
	}

	record, err := fetchRecord(bucketName, blobName)
	if err != nil {
		if err == storageclient.ErrBlobNotFound || err == storageclient.ErrBucketNotFound {
			return nil, ErrNotFound
		}
		sugar.Errorw("failed to fetch record process", "error", "err")
		return nil, ErrInternalServerError
	}
	fp, err := os.Open(record.GetPath())
	if err != nil {
		sugar.Errorw("failed to open recorded cache data", "error", "err")
		return nil, ErrInternalServerError
	}
	defer fp.Close()
	data, err := io.ReadAll(fp)
	if err != nil {
		sugar.Errorw("failed to read", "error", "err")
		return nil, ErrInternalServerError
	}

	contentType := record.ContentType
	if !slices.Contains(env.SUPPORTED_CONTENT_TYPES, contentType) {
		predicted, err := predictContentType(blobName)
		if err != nil {
			return nil, err
		}
		contentType = predicted
	}
	data, err = preprocess.ReduceImage(data, contentType, query)
	if err != nil {
		sugar.Errorw("failed to pre-processe object", "error", "err")
		return nil, ErrInternalServerError
	}
	product := &Product{
		Data:   data,
		Record: record,
	}
	return product, nil
}

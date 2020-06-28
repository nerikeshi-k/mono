package provider

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"mime"
	"os"
	"path"
	"time"

	"mono/config"
	"mono/env"
	"mono/preprocess"
	"mono/recordstore"
	"mono/storageclient"
	"mono/util"

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
	data, err := ioutil.ReadAll(fp)
	if err != nil {
		sugar.Errorw("failed to read", "error", "err")
		return nil, ErrInternalServerError
	}

	data, err = preprocess.ReduceImage(data, record.ContentType, query)
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

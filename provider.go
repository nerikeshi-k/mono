package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"mono/config"
	"mono/env"
	"mono/preprocess"
	"mono/recordstore"
	"mono/storageclient"
	"mono/utils"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func fetchRecord(bucketName string, blobName string) (*recordstore.Record, error) {
	sugar := zap.NewExample().Sugar()
	defer sugar.Sync()

	key := recordstore.GenerateKey(bucketName, blobName)
	record, err := recordstore.GetRecord(key)
	if err == nil && utils.DoesFileExist(record.GetPath()) {
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

func parsePreprocessQuery(c echo.Context) preprocess.Query {
	atoi := func(s string) uint {
		num, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			return 0
		}
		return uint(num)
	}
	query := preprocess.Query{
		MaxWidth:  atoi(c.QueryParam("maxwidth")),
		MaxHeight: atoi(c.QueryParam("maxheight")),
		Width:     atoi(c.QueryParam("width")),
		Height:    atoi(c.QueryParam("height")),
	}
	return query
}

func provide(c echo.Context) error {
	sugar := zap.NewExample().Sugar()
	defer sugar.Sync()

	// blob名取得
	blobName := c.QueryParam("name")
	if blobName == "" {
		return c.String(http.StatusBadRequest, fmt.Sprintf("invalid request"))
	}

	// bucket名取得
	bucketName := c.Request().Header.Get("X-Bucket-Name")
	if bucketName == "" {
		// 指定がないならconfigにある一個目のbuckets名とする
		buckets := config.Get().Buckets
		if len(buckets) > 0 {
			bucketName = config.Get().Buckets[0].Name
		}
	}

	// bucket名がconfig内に指定されているか確認
	bucketNameExists := false
	for _, bucket := range config.Get().Buckets {
		if bucket.Name == bucketName {
			bucketNameExists = true
			break
		}
	}
	if !bucketNameExists {
		return c.String(http.StatusBadRequest, fmt.Sprintf("bucket not found"))
	}

	record, err := fetchRecord(bucketName, blobName)
	if err != nil {
		if err == storageclient.ErrBlobNotFound || err == storageclient.ErrBucketNotFound {
			return c.String(http.StatusNotFound, fmt.Sprintf("not found"))
		}
		sugar.Errorw("failed to fetch record process", "error", "err")
		return c.String(http.StatusInternalServerError, fmt.Sprintf("server error"))
	}
	fp, err := os.Open(record.GetPath())
	if err != nil {
		sugar.Errorw("failed to open recorded cache data", "error", "err")
		return c.String(http.StatusInternalServerError, fmt.Sprintf("server error"))
	}
	defer fp.Close()
	data, err := ioutil.ReadAll(fp)
	if err != nil {
		sugar.Errorw("failed to read", "error", "err")
		return c.String(http.StatusInternalServerError, fmt.Sprintf("server error"))
	}

	q := parsePreprocessQuery(c)
	data, err = preprocess.ReduceImage(data, record.ContentType, q)
	if err != nil {
		sugar.Errorw("failed to pre-processe object", "error", "err")
		return c.String(http.StatusInternalServerError, fmt.Sprintf("server error"))
	}
	c.Response().Header().Set("Cache-Control", config.Get().CacheControlHeader)
	return c.Blob(http.StatusOK, record.ContentType, data)
}

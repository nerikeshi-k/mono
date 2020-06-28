package handler

import (
	"errors"
	"mono/config"
	"mono/preprocess"
	"mono/provider"
	"net/http"
	"net/url"
	"strings"

	echo "github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

var (
	// ErrInvalidRequest パラメータ不正
	ErrInvalidRequest = errors.New("invalid parameters")
)

// Query URLのパース結果
type Query struct {
	BlobName        string
	PreprocessQuery preprocess.Query
}

// Handle リクエストの受け口
func Handle(c echo.Context) error {
	sugar := zap.NewExample().Sugar()
	defer sugar.Sync()

	// URLパース
	query, err := parseURL(c.Request().URL)
	if err != nil {
		if err == ErrInvalidRequest {
			return c.String(http.StatusBadRequest, "invalid parameter")
		}
		sugar.Errorw("handle caught Internal Server Error", "error", err)
		return c.String(http.StatusInternalServerError, "server error")
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

	product, err := provider.Provide(bucketName, query.BlobName, query.PreprocessQuery)
	if err != nil {
		if err == provider.ErrNotFound {
			return c.String(http.StatusNotFound, "400 not found")
		}
		if err == provider.ErrInternalServerError {
			return c.String(http.StatusInternalServerError, "500 server error")
		}
		return c.String(http.StatusInternalServerError, "500 server error")
	}
	c.Response().Header().Set("Cache-Control", config.Get().CacheControlHeader)
	return c.Blob(http.StatusOK, product.Record.ContentType, product.Data)
}

func parseURL(URL *url.URL) (*Query, error) {
	sugar := zap.NewExample().Sugar()
	defer sugar.Sync()

	pathname := URL.Path[1:] // without first "/"
	i := strings.Index(pathname, "/")
	if i == -1 {
		return nil, ErrInvalidRequest
	}
	rawQuery := pathname[:i]
	blobName := pathname[i+1:]

	sugar.Debugw("values", "rawQuery", rawQuery, "blobName", blobName)
	preprocessQuery := parseRawQuery(rawQuery)
	query := &Query{
		BlobName:        blobName,
		PreprocessQuery: *preprocessQuery,
	}
	return query, nil
}

func parseRawQuery(raw string) *preprocess.Query {
	query := preprocess.Query{}
	fragments := strings.Split(raw, ",")
	for _, f := range fragments {
		i := strings.Index(f, "=")
		var key, value string
		if i == -1 {
			key = f
		} else {
			key = f[:i]
			value = f[i+1:]
		}
		insertKeyValue(&query, key, value)
	}
	return &query
}

func insertKeyValue(query *preprocess.Query, key string, value string) {
	switch key {
	case "w":
		query.MaxWidth = atoiPos(value)
	case "h":
		query.MaxHeight = atoiPos(value)
	case "wf":
		query.Width = atoiPos(value)
	case "hf":
		query.Height = atoiPos(value)
	}
}

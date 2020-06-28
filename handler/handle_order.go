package handler

import (
	"fmt"
	"mono/config"
	"mono/preprocess"
	"mono/provider"
	"net/http"
	"strconv"

	echo "github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// HandleOrder /order/パスのリクエストをさばく
// benienmaとの下位互換維持用
func HandleOrder(c echo.Context) error {
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

	q := parseOrderPreprocessQuery(c)

	product, err := provider.Provide(bucketName, blobName, q)
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

func parseOrderPreprocessQuery(c echo.Context) preprocess.Query {
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

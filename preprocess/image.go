package preprocess

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"

	"github.com/nfnt/resize"
)

// ReduceImage クエリに従って画像を加工して返す
func ReduceImage(data []byte, contentType string, q Query) ([]byte, error) {
	if contentType != "image/jpeg" && contentType != "image/png" {
		return data, nil
	}
	if q.IsDefault() {
		return data, nil
	}

	var img image.Image
	var err error
	switch contentType {
	case "image/jpeg":
		img, err = jpeg.Decode(bytes.NewReader(data))
	case "image/png":
		img, err = png.Decode(bytes.NewReader((data)))
	}
	if err != nil {
		return nil, err
	}

	// リサイズ
	if q.MaxWidth != 0 && q.MaxHeight != 0 {
		img = resize.Thumbnail(q.MaxWidth, q.MaxHeight, img, resize.Lanczos3)
	} else if q.MaxWidth != 0 && q.MaxHeight == 0 {
		img = resize.Thumbnail(q.MaxWidth, 4294967295, img, resize.Lanczos3)
	} else if q.MaxWidth == 0 && q.MaxHeight != 0 {
		img = resize.Thumbnail(4294967295, q.MaxHeight, img, resize.Lanczos3)
	} else if q.Width != 0 {
		img = resize.Resize(q.Width, 0, img, resize.Lanczos3)
	} else if q.Height != 0 {
		img = resize.Resize(0, q.Height, img, resize.Lanczos3)
	}

	buf := new(bytes.Buffer)
	switch contentType {
	case "image/jpeg":
		err = jpeg.Encode(buf, img, &jpeg.Options{Quality: 100})
	case "image/png":
		err = png.Encode(buf, img)
	}
	return buf.Bytes(), nil
}

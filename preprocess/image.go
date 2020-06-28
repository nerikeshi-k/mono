package preprocess

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"

	"github.com/disintegration/imaging"
)

// ReduceImage クエリに従って画像を加工して返す
func ReduceImage(data []byte, contentType string, q Query) ([]byte, error) {
	if contentType != "image/jpeg" && contentType != "image/png" {
		return data, nil
	}
	if q.HasNoMutation() {
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

	// リサイズ maxwidth, maxheightが指定されていた場合はwidth, heightより優先する
	size := img.Bounds().Size()
	if q.MaxWidth != 0 || q.MaxHeight != 0 {
		var w, h = q.MaxWidth, q.MaxHeight
		if w == 0 {
			w = size.X
		}
		if h == 0 {
			h = size.Y
		}
		img = imaging.Fit(img, w, h, imaging.Lanczos)
	} else if q.Width != 0 || q.Height != 0 {
		img = imaging.Resize(img, q.Width, q.Height, imaging.Lanczos)
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

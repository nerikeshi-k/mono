package preprocess

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/disintegration/imaging"
	"github.com/pixiv/go-libwebp/webp"
)

const MAX_WIDTH = 4096
const MAX_HEIGHT = 4096 * 2

func castToNRGBA(img image.Image) *image.NRGBA {
	conv, ok := img.(*image.NRGBA)
	if ok {
		return conv
	}
	return imaging.Clone(img)
}

func decode(data []byte, contentType string) (image.Image, error) {
	switch contentType {
	case "image/jpeg":
		decoded, err := jpeg.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		return castToNRGBA(decoded), nil
	case "image/png":
		decoded, err := png.Decode(bytes.NewReader((data)))
		if err != nil {
			return nil, err
		}
		return castToNRGBA(decoded), nil
	case "image/webp":
		bytes, err := io.ReadAll(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		decoded, err := webp.DecodeRGBA(bytes, &webp.DecoderOptions{})
		if err != nil {
			return nil, err
		}
		return castToNRGBA(decoded), nil
	default:
		return nil, fmt.Errorf("unsupported content type: %s", contentType)
	}
}

func encode(img image.Image, encodeTarget string) ([]byte, error) {
	buf := new(bytes.Buffer)
	switch encodeTarget {
	case "image/jpeg":
		err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 85})
		if err != nil {
			return nil, err
		}
	case "image/png":
		err := png.Encode(buf, img)
		if err != nil {
			return nil, err
		}
	case "image/webp":
		config, err := webp.ConfigPreset(webp.PresetDefault, 90)
		if err != nil {
			return nil, err
		}
		err = webp.EncodeRGBA(buf, img, config)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported content type: %s", encodeTarget)
	}
	return buf.Bytes(), nil
}

// ReduceImage クエリに従って画像を加工して返す
func ReduceImage(data []byte, sourceImageContentType string, q Query) ([]byte, error) {
	img, err := decode(data, sourceImageContentType)
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
	} else if (q.Width != 0 || q.Height != 0) && (q.Width <= MAX_WIDTH && q.Height <= MAX_HEIGHT) {
		img = imaging.Resize(img, q.Width, q.Height, imaging.Lanczos)
	}

	// 書き出し
	encodeTarget := q.EncodeTarget
	if encodeTarget == "" {
		encodeTarget = sourceImageContentType
	}
	return encode(img, encodeTarget)
}

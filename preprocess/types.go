package preprocess

// Query preprocess指示
type Query struct {
	MaxWidth     int    // 最大width
	MaxHeight    int    // 最大height
	Width        int    // width
	Height       int    // height
	EncodeTarget string // 出力時の形式 ("", "image/jpeg", "image/png", "image/webp")
}

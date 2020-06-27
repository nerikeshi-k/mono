package preprocess

// Query preprocess指示
type Query struct {
	MaxWidth  uint // 最大width
	MaxHeight uint // 最大height
	Width     uint // width
	Height    uint // height
}

// IsDefault 設定がない
func (q *Query) IsDefault() bool {
	if q.MaxWidth == 0 && q.MaxHeight == 0 &&
		q.Width == 0 && q.Height == 0 {
		return true
	}
	return false
}

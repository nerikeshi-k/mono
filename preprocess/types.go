package preprocess

// Query preprocess指示
type Query struct {
	MaxWidth  int // 最大width
	MaxHeight int // 最大height
	Width     int // width
	Height    int // height
}

// HasNoMutation 設定なし
func (q *Query) HasNoMutation() bool {
	if q.MaxWidth == 0 && q.MaxHeight == 0 &&
		q.Width == 0 && q.Height == 0 {
		return true
	}
	return false
}

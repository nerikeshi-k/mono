package recordstore

import (
	"encoding/json"
	"path"
	"time"

	"mono/config"
)

// Record storeに保存するデータ
type Record struct {
	BlobName        string    `json:"blob_name"`       // GCSのbucket内での名前
	CacheFileName   string    `json:"cache_file_name"` // キャッシュのファイル名 UUID
	Size            int64     `json:"size"`            // ファイルサイズ
	ContentType     string    `json:"content_type"`    // ContentType
	LastRequestedAt time.Time `json:"last_requested_at"`
	CreatedAt       time.Time `json:"created_at"`
}

// MarshalBinary Record -> json
func (r *Record) MarshalBinary() ([]byte, error) {
	b, err := json.Marshal(r)
	return b, err
}

// UnmarshalBinary json -> Record
func (r *Record) UnmarshalBinary(data []byte) error {
	err := json.Unmarshal(data, &r)
	return err
}

// GetPath キャッシュの実体のパスを返す
func (r *Record) GetPath() string {
	return path.Join(config.Get().CacheDirPath, r.CacheFileName)
}

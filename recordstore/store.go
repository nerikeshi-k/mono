package recordstore

import (
	"time"

	"crypto/md5"
	"encoding/hex"

	"github.com/nerikeshi-k/mono/config"
	"github.com/nerikeshi-k/mono/util"

	set "github.com/deckarep/golang-set"
	badger "github.com/dgraph-io/badger/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var (
	db *badger.DB
	// ErrRecordNotFound キーをもとにストアを探したがレコードがなかった
	ErrRecordNotFound = errors.New("record not found")
)

func init() {
	sugar := zap.NewExample().Sugar()
	defer sugar.Sync()

	var err error
	db, err = badger.Open(badger.DefaultOptions(config.Get().RecordStoreDirPath))
	if err != nil {
		sugar.Fatalw("Failed to open database", "error", err)
	}
}

// Close DBをクローズする
func Close() {
	db.Close()
}

// GenerateKey bucketNameとblobNameからKVSで使うkeyを作る
func GenerateKey(bucketName string, blobName string) string {
	joined := bucketName + "@" + blobName
	hasher := md5.New()
	hasher.Write([]byte(joined))
	return hex.EncodeToString(hasher.Sum(nil))
}

// GenerateCacheFileName UUIDを返すだけ
func GenerateCacheFileName() string {
	return util.GenerateUUID()
}

// GetRecord KVSからRecordを探して返す、なければnilとerrorを返す
func GetRecord(key string) (*Record, error) {
	sugar := zap.NewExample().Sugar()
	defer sugar.Sync()

	var data []byte
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		data, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrRecordNotFound
		}
		sugar.Errorw("failed to get item", "error", err)
		return nil, err
	}
	var record Record
	if err := record.UnmarshalBinary(data); err != nil {
		sugar.Errorw("failed to parse record", "error", err)
		return nil, err
	}
	return &record, nil
}

// SetRecord KVSにRecordをセットする
func SetRecord(key string, record *Record) error {
	sugar := zap.NewExample().Sugar()
	defer sugar.Sync()

	err := db.Update(func(txn *badger.Txn) error {
		bin, err := record.MarshalBinary()
		if err != nil {
			return err
		}
		entry := badger.NewEntry([]byte(key), bin).WithTTL(time.Second * time.Duration(config.Get().CacheExpires))
		err = txn.SetEntry(entry)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		sugar.Errorw("failed set record", "error", err)
		return err
	}
	return nil
}

// RunGC badgerのGCを走らせる
func RunGC() error {
	return db.RunValueLogGC(0.7)
}

// GetKeys 指定サイズ分のキーをiterateして返す
// size 0なら全て
func GetKeys(size int64) (set.Set, error) {
	var count int64
	keys := set.NewSet()

	err := db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Rewind(); it.Valid() && (size == 0 || count < size); it.Next() {
			if it.Item().IsDeletedOrExpired() {
				continue
			}
			count++
			keys.Add(string(it.Item().Key()))
		}
		return nil
	})
	return keys, err
}

// GetCacheFileNames 指定サイズ分のcache file nameを返す
// size 0なら全て
func GetCacheFileNames(size int64) (set.Set, error) {
	sugar := zap.NewExample().Sugar()
	defer sugar.Sync()

	var count int64
	names := set.NewSet()

	err := db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Rewind(); it.Valid() && (size == 0 || count < size); it.Next() {
			if it.Item().IsDeletedOrExpired() {
				continue
			}
			count++
			err := it.Item().Value(func(data []byte) error {
				record := Record{}
				err := record.UnmarshalBinary(data)
				if err != nil {
					return err
				}
				names.Add(record.CacheFileName)
				return nil
			})
			if err != nil {
				sugar.Errorw("failed to collect file names", "error", err)
			}
		}
		return nil
	})
	return names, err
}

// DeleteKeysWithSize size分だけキーを適当に消す
func DeleteKeysWithSize(size int64) error {
	keys, err := GetKeys(size)
	if err != nil {
		return err
	}

	err = db.Update(func(txn *badger.Txn) error {
		itr := keys.Iterator()
		for key := range itr.C {
			if ks, ok := key.(string); ok {
				txn.Delete([]byte(ks))
			}
		}
		return nil
	})
	return err
}

package gc

import (
	"os"
	"path"
	"time"

	"github.com/nerikeshi-k/mono/config"
	"github.com/nerikeshi-k/mono/env"
	"github.com/nerikeshi-k/mono/recordstore"
	"github.com/nerikeshi-k/mono/util"

	"go.uber.org/zap"
)

const randomDeleteBatchSize = 100
const badgerGCDuration = 5 * time.Minute

// Start 不要になったキャッシュとレコードの削除, badger GCの定期実行開始
func Start() {
	sugar := zap.NewExample().Sugar()
	defer sugar.Sync()
	if env.DEBUG {
		sugar.Debugw("start badger gc")
	}
	go startBadgerGC()
	processing := false
	ticker := time.NewTicker(time.Duration(config.Get().Collect.Span) * time.Second)
	defer ticker.Stop()
	if env.DEBUG {
		sugar.Debugw("start sweeper gc")
	}
	for range ticker.C {
		if !processing {
			processing = true
			sweep()
			processing = false
		}
	}
}

func startBadgerGC() {
	sugar := zap.NewExample().Sugar()
	defer sugar.Sync()

	ticker := time.NewTicker(badgerGCDuration)
	defer ticker.Stop()
	for range ticker.C {
	again:
		if env.DEBUG {
			sugar.Debugw("badger GC started")
		}
		err := recordstore.RunGC()
		if err == nil {
			goto again
		}
		if env.DEBUG {
			sugar.Debugw("badger GC finished")
		}
	}
}

func sweep() {
	sugar := zap.NewExample().Sugar()
	defer sugar.Sync()

	if env.DEBUG {
		sugar.Debugw("sweep started")
	}
	sweepRecordsIfVolumeNealyFull()
	unreachableCacheFileNames := collectUnreachableCacheFileNames()
	if env.DEBUG {
		sugar.Debugw("unreachable file names collected", "names", unreachableCacheFileNames)
	}
	for _, name := range unreachableCacheFileNames {
		p := path.Join(config.Get().CacheDirPath, name)
		if err := os.Remove(p); err != nil {
			sugar.Errorw("failed to remove file", "name", name, "error", err)
		}
	}
	if env.DEBUG {
		sugar.Debugw("sweep finised")
	}
}

// recordから参照されていないキャッシュファイルの名前を返す
func collectUnreachableCacheFileNames() []string {
	sugar := zap.NewExample().Sugar()
	defer sugar.Sync()

	unreachables := []string{}
	recordedNames, err := recordstore.GetCacheFileNames(0)
	if err != nil {
		sugar.Errorw("failed to get keys on collectUnreachables", "error", err)
		return []string{}
	}
	cacheFileNames, err := util.ListDir(config.Get().CacheDirPath)
	if err != nil {
		sugar.Errorw("failed to get cache file names on collectUnreachables", "error", err)
		return []string{}
	}
	for _, name := range cacheFileNames {
		if !recordedNames.Contains(name) {
			unreachables = append(unreachables, name)
		}
	}
	return unreachables
}

// キャッシュ用ディレクトリの容量が限界に近くなってきた場合キーをいくつか消す
func sweepRecordsIfVolumeNealyFull() error {
	volume := util.GetDirSizeMB(config.Get().CacheDirPath)
	// if volume is over 90% of max
	if volume > (float64(config.Get().MaxCacheVolume) * 0.9) {
		recordstore.DeleteKeysWithSize(randomDeleteBatchSize)
	}
	return nil
}

package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"

	"mono/utils"
)

var config Config

// Config 設定ファイル
type Config struct {
	Port               int64  `json:"port"`
	CacheDirPath       string `json:"cache_volume_path"`
	CacheControlHeader string `json:"cache_control_header"`
	RecordStoreDirPath string `json:"record_store_volume_path"`
	CacheExpires       int64  `json:"cache_expires"`
	MaxCacheVolume     int64  `json:"max_cache_volume"`
	Buckets            []struct {
		Name string `json:"name"`
	} `json:"buckets"`
	Collect struct {
		Span int64 `json:"span"`
	} `json:"collect"`
}

func init() {
	err := Load()
	if err != nil {
		panic(err)
	}
}

// Load 設定ファイルを読み込んでconfigに与える
func Load() error {
	var configFilePath string
	flag.StringVar(&configFilePath, "conf", "", "path of config.json")
	flag.Parse()
	if configFilePath == "" {
		return fmt.Errorf("use --conf to set config.json path")
	}
	if !utils.DoesFileExist(configFilePath) {
		return fmt.Errorf("%s does not exist", configFilePath)
	}
	bytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bytes, &config); err != nil {
		return err
	}
	if !utils.DoesFileExist(config.CacheDirPath) {
		return fmt.Errorf("Object caching dir %s does not exist", config.CacheDirPath)
	}
	return nil
}

// Get 設定構造体を返却する
func Get() Config {
	return config
}

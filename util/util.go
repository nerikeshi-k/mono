package util

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// DoesFileExist ファイルの存在を調べる。あればtrue, なければfalse
func DoesFileExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// ListDir 与えられたパスのディレクトリのファイル一覧を返す
func ListDir(dir string) ([]string, error) {
	filenames := []string{}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return filenames, err
	}
	for _, file := range files {
		filenames = append(filenames, file.Name())
	}
	return filenames, nil
}

// GetDirSizeMB 指定ディレクトリのサイズをMB単位で返す
func GetDirSizeMB(path string) float64 {
	// from https://stackoverflow.com/questions/32482673/golang-how-to-get-directory-total-size
	var dirSize int64
	readSize := func(path string, file os.FileInfo, err error) error {
		if !file.IsDir() {
			dirSize += file.Size()
		}
		return nil
	}
	filepath.Walk(path, readSize)
	sizeMB := float64(dirSize) / 1024.0 / 1024.0
	return sizeMB
}

// GenerateUUID UUID文字列を返す
func GenerateUUID() string {
	uuid := uuid.New()
	return uuid.String()
}

// GetExtension contentType -> Ext
func GetExtension(contentType string) string {
	switch contentType {
	case "image/png":
		return "png"
	case "image/jpeg":
		return "jpeg"
	}
	return ""
}

package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"strings"
)

// 获取文件MD5码
func GetFileMD5(fileName string) (string, error) {
	fileName = strings.TrimSpace(fileName)
	if fileName == "" {
		return "", nil
	}
	f, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)
	md5Encoder := md5.New()
	_, err = io.Copy(md5Encoder, f)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(md5Encoder.Sum(nil)), nil
}

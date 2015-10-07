package glweb

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"hash"
	"os"
	"path/filepath"
	"strings"
)

func WalkDir(dirPth, suffix string) (files []string, err error) {
	files = make([]string, 0, 30)
	suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写
	path, err := filepath.Abs(dirPth)
	if err != nil {
		return nil, errors.New("File Abs Path is missing!" + path)
	}
	err = filepath.Walk(path, func(filename string, fi os.FileInfo, err error) error { //遍历目录
		//if err != nil { //忽略错误
		// return err
		//}
		if fi.IsDir() { // 忽略目录
			return nil
		}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) {
			files = append(files, filename)
		}
		return nil
	})
	return files, err
}

var _md5 hash.Hash

func MD5(value string) string {
	if _md5 == nil {
		_md5 = md5.New()
	}
	_md5.Reset()
	_md5.Write([]byte(value)) // 需要加密的字符串为 123456
	return hex.EncodeToString(_md5.Sum(nil))
}

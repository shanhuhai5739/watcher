package utils

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"os"
	"unsafe"
)

var (
	ErrEmptyArguments = errors.New("Argument Cann't be empty")
)

/*
判断文件是否存在，该函数和PHP的file_exists一致，当filename 为 文件/文件夹/符号链接 时均返回 true
*/
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func FileWrite(filename string, content *string) (err error) {
	fd, err := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	defer fd.Close()
	bw := bufio.NewWriter(fd)

	_, err = bw.WriteString(*content)
	if err != nil {
		return
	}
	bw.Flush()

	return
}

/*
判断给定的filename是否是一个目录
*/
func IsDir(filename string) (bool, error) {

	if len(filename) <= 0 {
		return false, ErrEmptyArguments
	}

	stat, err := os.Stat(filename)
	if err != nil {
		return false, err
	}

	if !stat.IsDir() {
		return false, nil
	}

	return true, nil
}

func LoadFile(filePath string) (string, error) {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return Bytes2Str(b), nil
}

func Bytes2Str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

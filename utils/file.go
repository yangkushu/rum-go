package utils

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 将文件名分解为路径、名、后缀
// 如：/dir/filename.ext，分解结果为：/dir, filename, ext
func SplitFilename(filename string) (dir, name, ext string) {
	dir = filepath.Dir(filename)

	baseName := filepath.Base(filename)
	ext = filepath.Ext(baseName)
	name = baseName[:len(baseName)-len(ext)]
	if len(ext) > 0 {
		ext = ext[1:]
	}
	return
}

func TouchFile(file string) error {
	if !FileExist(file) {
		if fd, err := os.Create(file); err != nil {
			return err
		} else {
			_ = fd.Close()
		}
		return nil
	}
	now := time.Now()

	return os.Chtimes(file, now, now)
}

// 文件不存在返回 -1,nil
func GetFileModifyDay(file string) (int, error) {
	fileInfo, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			return -1, nil
		} else {
			return 0, err
		}
	}

	return fileInfo.ModTime().Day(), nil
}

func FileExist(filename string) bool {
	if info, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	} else if info.IsDir() {
		return false
	} else {
		return true
	}
}

// 获取文件大小
func GetFileSize(filename string) (int64, error) {
	if fileInfo, err := os.Stat(filename); nil != err {
		return 0, err
	} else if fileInfo.IsDir() {
		return 0, errors.New("not mormal file")
	} else {
		return fileInfo.Size(), nil
	}
}

// 把缓冲区数据保存成文件
func SaveDataToFile(filename string, data []byte) error {
	// 创建一个新的文件
	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer func() {
		if err := file.Close(); nil != err {
			fmt.Println("file.Close fail", err.Error())
		}
	}()

	// 创建一个缓冲区
	buf := bytes.NewBuffer(data)

	// 将缓冲区的内容写入文件
	_, err = buf.WriteTo(file)
	if err != nil {
		return err
	}

	return nil
}

// 替换文件名后缀
func ChangeSuffix(filename, suffix string) string {
	if len(filename) <= 0 {
		return ""
	}

	if index := strings.LastIndex(filename, "."); index >= 0 {
		return filename[:index] + suffix
	} else {
		return filename + suffix
	}
}

// 文件名后附件一个字符串：abc.txt + xyz = abcxyz.txt
func AddStringToFilename(filename, add string) string {
	if index := strings.LastIndex(filename, "."); index >= 0 {
		return filename[:index] + add + filename[index:]
	} else {
		return filename + add
	}
}

package utils

import (
	"errors"
	jsoniter "github.com/json-iterator/go"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Byte uint64

// 文件大小，单位
const (
	B  = Byte(1)
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
	TB = 1024 * GB
	PB = 1024 * TB
	EB = 1024 * PB
	//ZB = 1024 * EB
	//YB = 1024 * ZB
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

/*
检查目录是否存在，不存在则创建
*/
func CheckCreateDir(dir_path string) error {
	ok, err := PathExists(dir_path)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	return Mkdir(dir_path)
}

/*
判断一个路径的文件是否存在
*/
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

/*
递归创建目录
*/
func Mkdir(path string) error {
	err := os.MkdirAll(path, os.ModePerm)
	//	err := os.Mkdir(path, os.ModeDir)
	if err != nil {
		//		fmt.Println("创建文件夹失败", path, err)
		return err
	}
	return nil
}

/*
保存对象为json格式
*/
func SaveJsonFile(name string, o interface{}) error {
	bs, err := json.Marshal(o)
	if err != nil {
		return err
	}
	return SaveFile(name, &bs)
}

/*
保存文件
保存文件步骤：
1.创建临时文件
2.删除原文件
3.修改临时文件名称为原文件名称
*/
func SaveFile(name string, bs *[]byte) error {
	//创建临时文件
	now := strconv.Itoa(int(time.Now().Unix()))
	tempname := name + "." + now
	file, err := os.OpenFile(tempname, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		file.Close()
		return err
	}
	_, err = file.Write(*bs)
	if err != nil {
		file.Close()
		return err
	}
	file.Close()
	//删除旧文件
	ok, err := PathExists(name)
	if err != nil {
		// utils.Log.Info().Msgf("删除旧文件失败", err)
		return err
	}
	if ok {
		err = os.Remove(name)
		if err != nil {
			return err
		}
	}

	//重命名文件
	err = os.Rename(tempname, name)
	if err != nil {
		return err
	}
	return nil
}

/*
覆盖原文件
1.删除原文件
2.并将临时文件重命名为原文件
3.当只存在临时文件时，将临时文件命名为原文件
*/
func RenameTempFile(name string) error {
	ok, err := PathExists(name)
	if err != nil {
		return err
	}

	files, err := filepath.Glob(name + ".*")
	if err != nil {
		return err
	}

	if ok && len(files) > 0 { //临时文件和原文件同时存在则删除临时文件
		for _, f := range files {
			err = os.Remove(f)
			if err != nil {
				return err
			}
		}
	} else if len(files) > 0 { //只存在在临时文件时将临时文件命名成原文件
		tempname := files[0]
		err = os.Rename(tempname, name)
		if err != nil {
			return err
		}
	}
	return nil
}

/*
把文件全路径切分成子路径
*/
func FilePathSplit(path string) []string {
	names := make([]string, 0)
	// var beforName = path
	var dirName = path
	var fileName string
	for i := 0; i < 10; i++ {
		_, fileName = filepath.Split(dirName)
		if fileName == "" {
			names = append(names, dirName)
			break
		}
		dirName = filepath.Dir(dirName)
		names = append(names, fileName)
	}

	myReverse(names)
	return names
}

func myReverse(l []string) {
	for i := 0; i < int(len(l)/2); i++ {
		li := len(l) - i - 1
		l[i], l[li] = l[li], l[i]
	}
}

/*
读取文件前512字节，判断文件类型
*/
func FileContentType(path string) (string, os.FileInfo, error) {
	file, err := os.Open(path) // 打开文件
	if err != nil {
		return "", nil, err
	}
	defer file.Close() // 确保文件在函数结束时关闭
	size := 512
	fileinfo, err := file.Stat()
	if err != nil {
		return "", nil, err
	}
	//如果文件太小
	if fileinfo.Size() < int64(size) {
		return "", fileinfo, nil
	}

	buffer := make([]byte, size)
	n, err := io.ReadFull(file, buffer)
	if err != nil {
		return "", fileinfo, err
	}
	if n != size {
		return "", fileinfo, errors.New("file size error")
	}
	mimeType := http.DetectContentType(buffer)
	return mimeType, fileinfo, nil
}

/*
读取文件前512字节，判断文件类型
*/
func FileContentTypeIsImage(path string) (bool, string, os.FileInfo, error) {
	mimeType, fileinfo, err := FileContentType(path)
	if err != nil {
		return false, "", fileinfo, err
	}
	if len(mimeType) < 5 {
		return false, mimeType, fileinfo, nil
	}
	if mimeType[:5] == "image" {
		return true, mimeType, fileinfo, nil
	}
	return false, mimeType, fileinfo, nil
}

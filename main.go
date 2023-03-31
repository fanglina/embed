package main

import (
	"bufio"
	"embed"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

//这一行是go build的时候把conf.ini打包到.exe中，不能去掉
//go:embed conf.ini
var ConfStr string

//go:embed temp
var tempFS embed.FS

func main() {
	fmt.Println(ConfStr)
	rejectConf(ConfStr)
	EjectDir(tempFS, "temp")
}

// 把读出来的文件写回到文件中， 为了可以修改配置把他写回去
func rejectConf(data string) {
	fileName := "conf.ini"
	isExist := fileExists(fileName)
	if  isExist {
		return
	}
	//打开一个文件
	fileHandle, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer fileHandle.Close()
	buff := bufio.NewWriterSize(fileHandle, len(data))
	_, err = buff.WriteString(data)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	//清缓存
	err = buff.Flush()
	return
}

// 释放整个目录
func EjectDir(licenceFs embed.FS, dir string ) {
	// 初始化静态资源
	embedFS, err := fs.Sub(licenceFs, dir)
	if err != nil {
		fmt.Printf("无法初始化License, %s", err)
		return
	}

	var walk func(relPath string, d fs.DirEntry, err error) error
	walk = func(relPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return errors.Errorf("无法获取[%s]的信息, %s, 跳过...", relPath, err)
		}

		if !d.IsDir() {
			// 写入文件
			out, err := CreatNestedFile(filepath.Join(RelativePath(""), dir, relPath))
			defer out.Close()

			if err != nil {
				return errors.Errorf("无法创建文件[%s], %s, 跳过...", relPath, err)
			}

			obj, _ := embedFS.Open(relPath)
			if _, err := io.Copy(out, bufio.NewReader(obj)); err != nil {
				return errors.Errorf("无法写入文件[%s], %s, 跳过...", relPath, err)
			}
		}
		return nil
	}

	// util.Log().Info("开始导出内置静态资源...")
	err = fs.WalkDir(embedFS, ".", walk)
	if err != nil {
		fmt.Printf("导出内置静态资源遇到错误：%s", err)
		return
	}
	fmt.Println("内置静态资源导出完成")
}



// 判断所给路径文件/文件夹是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		fmt.Println(err)
		return false
	}
	return true
}


// CreatNestedFile 给定path创建文件，如果目录不存在就递归创建
func CreatNestedFile(path string) (*os.File, error) {
	basePath := filepath.Dir(path)
	if !fileExists(basePath) {
		err := os.MkdirAll(basePath, 0700)
		if err != nil {
			return nil, err
		}
	}

	return os.Create(path)
}

// RelativePath 获取相对可执行文件的路径
func RelativePath(name string) string {
	if filepath.IsAbs(name) {
		return name
	}
	e, _ := os.Executable()
	return filepath.Join(filepath.Dir(e), name)
}

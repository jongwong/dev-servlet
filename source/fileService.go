package main

import (
	"io"
	"log"
	"os"
	"path"
)

type Files struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func checkFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}
func pathResolve(path string) string {
	s3 := *workPath + path
	return s3
}

func writeStringToFile(filepath, content string) {
	log.Println(filepath)
	var f *os.File
	var err1 error

	var dir, _ = path.Split(filepath)
	if !checkFileIsExist(dir) {
		err1 = os.MkdirAll(dir, os.ModePerm)
		throwError(err1)
	}

	f, err1 = os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY, 0666)
	throwError(err1)

	_, err1 = io.WriteString(f, content) //写入文件(字符串)
	throwError(err1)
}
func writeFiles(files []Files) {
	for i := range files {
		var file = files[i]
		writeStringToFile(pathResolve(file.Path), file.Content)
	}
}

func deleteFile(path string) {
	err := os.RemoveAll(path)
	throwError(err)
}
func deleteFiles(files []Files) {
	for i := range files {
		var file = files[i]
		deleteFile(pathResolve(file.Path))

	}
}

func clearAndWriteFiles(files []Files) {
	deleteFile(*workPath)
}

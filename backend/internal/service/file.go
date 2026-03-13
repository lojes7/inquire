package service

import (
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func saveFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

func getFileType(fileName string) string {
	file, err := os.Open(fileName)
	if err != nil {
		log.Printf("无法打开文件 %s: %v", fileName, err)
		return "application/octet-stream"
	}
	defer file.Close()

	// 读取前 512 字节用于检测
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		log.Printf("读取文件头失败: %v", err)
		return "application/octet-stream"
	}

	// 基于内容检测
	contentType := http.DetectContentType(buffer[:n])

	// 特殊情况：DetectContentType 对 .txt 返回 application/octet-stream
	// 可手动修正
	if contentType == "application/octet-stream" {
		ext := strings.ToLower(filepath.Ext(fileName))
		if ext == ".txt" || ext == ".log" || ext == ".csv" {
			return "text/plain"
		}
	}

	return contentType
}

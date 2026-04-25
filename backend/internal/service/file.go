package service

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/lojes7/inquire/pkg/infra"
	"github.com/lojes7/inquire/pkg/secure"
	"github.com/lojes7/inquire/pkg/utils"
)

const saveFileMaxRetry = 2

type SavedFileInfo struct {
	OriginalFileName string
	FileName         string
	StoredFileName   string
	FilePath         string
	FileType         string
	FileSize         int64
	FileExt          string
}

func sanitizeFilename(fileName string) string {
	cleaned := strings.TrimSpace(fileName)
	cleaned = strings.ReplaceAll(cleaned, "\\", "/")
	cleaned = filepath.Base(cleaned)
	cleaned = strings.TrimSpace(cleaned)

	if cleaned == "" || cleaned == "." || cleaned == "/" {
		return ""
	}

	return cleaned
}

func SaveFileIntoServer(file *multipart.FileHeader) (*SavedFileInfo, error) {
	if file == nil {
		return nil, secure.Wrap(400, "文件不能为空", errors.New("nil file header"))
	}

	uploadDir := infra.GetFilePath()
	if strings.TrimSpace(uploadDir) == "" {
		return nil, secure.Wrap(500, "文件目录配置错误", errors.New("empty file storage path"))
	}

	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		log.Println(err)
		return nil, secure.Wrap(500, "创建文件目录失败", err)
	}

	originalFileName := sanitizeFilename(file.Filename)
	if originalFileName == "" {
		originalFileName = fmt.Sprintf("unnamed_%d", utils.NewUniqueID())
	}

	rawExt := filepath.Ext(originalFileName)
	ext := strings.ToLower(rawExt)
	fileName := strings.TrimSuffix(originalFileName, rawExt)
	if fileName == "" {
		fileName = originalFileName
	}

	for i := 0; i < saveFileMaxRetry; i++ {
		storedFileName := fmt.Sprintf("%d%s", utils.NewUniqueID(), ext)
		filePath := filepath.Join(uploadDir, storedFileName)

		if err := saveFile(file, filePath); err != nil {
			if errors.Is(err, os.ErrExist) {
				continue
			}
			log.Println(err)
			return nil, secure.Wrap(500, "保存文件失败", err)
		}

		return &SavedFileInfo{
			OriginalFileName: originalFileName,
			FileName:         fileName,
			StoredFileName:   storedFileName,
			FilePath:         filePath,
			FileType:         getFileType(filePath),
			FileSize:         file.Size,
			FileExt:          ext,
		}, nil
	}

	return nil, secure.Wrap(500, "保存文件失败", errors.New("文件名冲突重试失败"))
}

func saveFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, src); err != nil {
		_ = os.Remove(dst)
		return err
	}

	return nil
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

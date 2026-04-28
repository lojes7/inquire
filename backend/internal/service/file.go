package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgconn"
	"github.com/lojes7/inquire/internal/model"
	"github.com/lojes7/inquire/pkg/infra"
	"github.com/lojes7/inquire/pkg/secure"
	"github.com/lojes7/inquire/pkg/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

// sanitizeFilename 进行基本的文件名清理，防止路径穿越和非法字符。
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

// SaveFileIntoServer 负责将上传的文件保存到服务器磁盘，并返回保存后的文件信息。
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

// saveFile SaveFileIntoServer 的 helper-func
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

// getFileType 基于文件内容检测 MIME 类型，fallback 到 application/octet-stream。
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

// computeFileHash 计算上传文件的 SHA-256 哈希值，用于后续的去重逻辑。
func computeFileHash(file *multipart.FileHeader) (string, error) {
	if file == nil {
		return "", secure.Wrap(400, "文件不能为空", errors.New("nil file header"))
	}

	// 这里使用流式读取来计算 SHA-256：
	// 1) 不会把整个文件一次性加载到内存，适合大文件场景；
	// 2) 多次调用 file.Open() 会拿到新的 reader，不会影响后续保存文件的读取位置。
	src, err := file.Open()
	if err != nil {
		log.Println(err)
		return "", secure.Wrap(500, "读取文件失败", err)
	}
	defer src.Close()

	hasher := sha256.New()
	// io.Copy 会把文件内容持续喂给哈希器，最终得到固定 32 字节摘要。
	if _, err = io.Copy(hasher, src); err != nil {
		log.Println(err)
		return "", secure.Wrap(500, "计算文件哈希失败", err)
	}

	// 以十六进制字符串返回，长度固定 64，与数据库 char(64) 完全匹配。
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// getExistingFileByHash 根据文件哈希查询数据库，判断是否已有相同内容的文件记录。
func getExistingFileByHash(tx *gorm.DB, hashValue string) (*model.File, error) {
	var file model.File
	// model.File 使用了 gorm 软删除，默认查询会自动附带 deleted_at IS NULL，
	// 因此这里天然只会命中“未软删除”的去重记录。
	err := tx.Model(&model.File{}).
		Where("hash_value = ?", hashValue).
		First(&file).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		log.Println(err)
		return nil, secure.Wrap(500, "查询文件哈希失败", err)
	}

	return &file, nil
}

// isFileHashUniqueViolation 判断数据库错误是否由 files.hash_value 唯一索引冲突引起。
func isFileHashUniqueViolation(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	// 只识别 files.hash_value 对应的唯一索引冲突，避免把其他唯一键冲突误判成“重复文件”。
	return pgErr.Code == "23505" && strings.Contains(pgErr.ConstraintName, fileHashUniqueIndexName)
}

// createFileRecord 在数据库中创建新的文件记录，返回创建的记录或特定错误。
func createFileRecord(tx *gorm.DB, hashValue string, savedFileInfo *SavedFileInfo) (*model.File, error) {
	if savedFileInfo == nil {
		return nil, secure.Wrap(500, "创建文件记录失败", errors.New("saved file info is nil"))
	}

	fileName := strings.TrimSpace(savedFileInfo.FileName)
	if fileName == "" {
		fileName = strings.TrimSpace(savedFileInfo.OriginalFileName)
	}
	if fileName == "" {
		fileName = fmt.Sprintf("unnamed_%d", utils.NewUniqueID())
	}

	fileType := strings.TrimSpace(savedFileInfo.FileType)
	if fileType == "" {
		fileType = "application/octet-stream"
	}

	filePath := strings.TrimSpace(savedFileInfo.FilePath)
	if filePath == "" {
		return nil, secure.Wrap(500, "创建文件记录失败", errors.New("file path is empty"))
	}

	newFile := model.File{
		FileName:  fileName,
		FileType:  fileType,
		FileURL:   filePath,
		FileSize:  savedFileInfo.FileSize,
		HashValue: hashValue,
	}

	if err := tx.Create(&newFile).Error; err != nil {
		// 并发上传同一文件时，可能会命中 hash_value 唯一索引。
		// 这里返回哨兵错误，交给上层执行“回查并复用”的补偿流程。
		if isFileHashUniqueViolation(err) {
			return nil, errFileHashConflict
		}
		log.Println(err)
		return nil, secure.Wrap(500, "创建文件记录失败", err)
	}

	return &newFile, nil
}

func checkFileByHash(tx *gorm.DB, hashValue string) (*model.File, bool, error) {
	// 只做查询，不在这里创建记录。
	// 这样 SendFile 就能采用“先查重 -> 若不存在则先落盘 -> 再写 files”的顺序。
	existingFile, err := getExistingFileByHash(tx, hashValue)
	if err != nil {
		return nil, false, err
	}
	if existingFile != nil {
		return existingFile, true, nil
	}

	return nil, false, nil
}

// createUserFileRelations SendFile 的辅助函数。
//
// 设计说明：
// 1. 这里不做应用层“先查后写”的去重判断，避免并发窗口导致判断失真。
// 2. 关系唯一性完全交给数据库唯一索引(user_id, file_id, deleted_at is null)保证。
// 3. 批量写入时使用 ON CONFLICT DO NOTHING，若命中重复关系则静默跳过，不影响发送流程。
func createUserFileRelations(tx *gorm.DB, conversationID, fileID uint64) error {
	var userIDs []uint64
	err := tx.Model(&model.ConversationUser{}).
		Where("conversation_id = ? AND deleted_at IS NULL", conversationID).
		Pluck("user_id", &userIDs).Error
	if err != nil {
		log.Println(err)
		return secure.Wrap(500, "查询会话成员失败", err)
	}

	if len(userIDs) == 0 {
		return secure.Wrap(500, "会话成员为空", errors.New("conversation has no members"))
	}

	userFiles := make([]model.UserFile, 0, len(userIDs))
	for _, userID := range userIDs {
		userFiles = append(userFiles, model.UserFile{
			UserID: userID,
			FileID: fileID,
		})
	}

	if len(userFiles) == 0 {
		return nil
	}

	if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&userFiles).Error; err != nil {
		log.Println(err)
		return secure.Wrap(500, "写入 user_files 关系失败", err)
	}

	return nil
}

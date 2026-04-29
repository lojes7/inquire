package service

import (
	"errors"
	"log"
	"mime/multipart"
	"os"
	"strings"

	"github.com/lojes7/inquire/internal/model"
	"github.com/lojes7/inquire/pkg/infra"
	"github.com/lojes7/inquire/pkg/secure"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UploadFile 上传文件到用户工作区。
// 流程：落盘 → 哈希去重 → 写 files 表 → 写 user_files 关系 → 异步触发嵌入。
func UploadFile(userID uint64, file *multipart.FileHeader) (*model.UploadFileResp, error) {
	if file == nil {
		return nil, secure.Wrap(400, "文件不能为空", errors.New("nil file header"))
	}

	// 1) 计算哈希用于去重
	hashValue, err := computeFileHash(file)
	if err != nil {
		return nil, err
	}

	db := infra.GetDB()
	var fileRecord *model.File
	var savedFilePath string // 事务失败时用于清理孤儿文件

	err = db.Transaction(func(tx *gorm.DB) error {
		// 2) 按 Hash 查重
		existing, existed, err := checkFileByHash(tx, hashValue)
		if err != nil {
			return err
		}
		if existed {
			fileRecord = existing
		} else {
			// 3) 落盘 + 写 files 记录
			saved, saveErr := SaveFileIntoServer(file)
			if saveErr != nil {
				return saveErr
			}
			savedFilePath = saved.FilePath

			fileRecord, err = createFileRecord(tx, hashValue, saved)
			if err != nil {
				if errors.Is(err, errFileHashConflict) {
					// 并发补偿
					existingFile, existedErr := getExistingFileByHash(tx, hashValue)
					if existedErr != nil {
						return existedErr
					}
					if existingFile == nil {
						return secure.Wrap(500, "文件重复校验失败", errors.New("hash conflict but file not found"))
					}
					if removeErr := os.Remove(saved.FilePath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
						log.Printf("并发冲突后清理文件失败: %v", removeErr)
					}
					fileRecord = existingFile
				} else {
					return err
				}
			}
		}

		// 4) 建立 user_files 关系（唯一索引 ON CONFLICT DO NOTHING）
		uf := model.UserFile{
			UserID: userID,
			FileID: fileRecord.ID,
		}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&uf).Error; err != nil {
			log.Println(err)
			return secure.Wrap(500, "创建用户文件关系失败", err)
		}

		return nil
	})

	if err != nil {
		// 事务失败时清理刚落盘的文件，避免产生孤儿文件
		if strings.TrimSpace(savedFilePath) != "" {
			if removeErr := os.Remove(savedFilePath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
				log.Printf("事务失败后清理文件失败: %v", removeErr)
			}
		}
		return nil, err
	}

	// 5) 非阻塞触发向量化
	if strings.TrimSpace(fileRecord.FileURL) != "" {
		go EmbedFile(fileRecord.ID, fileRecord.FileURL, fileRecord.FileType)
	}

	return &model.UploadFileResp{
		FileInfo: model.FileInfoResp{
			FileID:   fileRecord.ID,
			FileName: fileRecord.FileName,
			FileSize: fileRecord.FileSize,
			FileType: fileRecord.FileType,
		},
	}, nil
}

// WorkspaceFileList 获取用户工作区中的所有文件。
func WorkspaceFileList(userID uint64) ([]model.FileInfoResp, error) {
	db := infra.GetDB()

	var files []model.FileInfoResp
	err := db.Model(&model.UserFile{}).
		Select("f.id AS file_id, f.file_name, f.file_size, f.file_type").
		Joins("JOIN files f ON f.id = user_files.file_id AND f.deleted_at IS NULL").
		Where("user_files.user_id = ? AND user_files.deleted_at IS NULL", userID).
		Order("user_files.created_at DESC").
		Scan(&files).Error
	if err != nil {
		log.Println(err)
		return nil, secure.Wrap(500, "查询工作区文件失败", err)
	}

	if files == nil {
		files = make([]model.FileInfoResp, 0)
	}

	return files, nil
}

func DownloadFile(userID, messageID uint64) (string, error) {
	db := infra.GetDB()

	// 一次查询完成：消息存在 + 用户在对话中 + 文件存在
	var file model.File
	err := db.Model(&model.File{}).
		Select("files.file_url").
		Joins("JOIN messages m ON m.file_id = files.id").
		Joins("JOIN conversation_users cu ON cu.conversation_id = m.conversation_id").
		Where("cu.user_id = ? AND m.status = ?",
			userID, model.FILE).
		First(&file).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", secure.Wrap(403, "文件不存在或无访问权限", err)
		}
		log.Println("DB error:", err)
		return "", secure.Wrap(500, "查询文件失败", err)
	}

	return file.FileURL, nil
}

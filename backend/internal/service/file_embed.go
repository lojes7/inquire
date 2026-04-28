package service

import (
	"fmt"
	"log"
	"strings"

	"github.com/lojes7/inquire/internal/model"
	"github.com/lojes7/inquire/pkg/infra"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// EmbedFile 处理文件的向量嵌入管线：
// 1. 幂等检查 — 若已有向量则跳过
// 2. 调用 AI 服务生成向量
// 3. 事务批量写入 file_vectors 表
//
// 该函数设计为在 goroutine 中异步调用，不阻塞主流程。
func EmbedFile(fileID uint64, filePath string, fileType string) {
	log.Printf("[file-embed] Start processing file_id=%d path=%s type=%s", fileID, filePath, fileType)

	// 幂等检查：若该文件已有向量记录则跳过
	if alreadyEmbedded(fileID) {
		log.Printf("[file-embed] File %d already has vectors, skipping", fileID)
		return
	}

	// 调用 AI 服务生成向量
	resp, err := callAIEmbed(filePath, fileID, fileType)
	if err != nil {
		log.Printf("[file-embed] AI embedding failed for file_id=%d: %v", fileID, err)
		return
	}

	log.Printf("[file-embed] Received %d vectors for file_id=%d (chunk_count=%d)",
		len(resp.Vectors), fileID, resp.ChunkCount)

	// 校验返回数据：确保所有向量的维度正确
	for _, item := range resp.Vectors {
		if len(item.Vector) != 1024 {
			log.Printf("[file-embed] Vector dimension mismatch for file_id=%d chunk %d: got %d, expected 1024",
				fileID, item.Number, len(item.Vector))
			return
		}
	}

	// 事务写入 file_vectors
	if err := persistVectors(fileID, resp.Vectors); err != nil {
		log.Printf("[file-embed] Failed to persist vectors for file_id=%d: %v", fileID, err)
		return
	}

	log.Printf("[file-embed] Successfully persisted %d vectors for file_id=%d", len(resp.Vectors), fileID)
}

// alreadyEmbedded 检查 file_vectors 表中是否已有该文件的向量记录。
func alreadyEmbedded(fileID uint64) bool {
	db := infra.GetDB()
	var count int64
	if err := db.Model(&model.FileVector{}).
		Where("file_id = ?", fileID).
		Count(&count).Error; err != nil {
		log.Printf("[file-embed] Idempotency check failed for file_id=%d: %v", fileID, err)
		// 查询失败时保守处理：不跳过，继续尝试写入
		return false
	}
	return count > 0
}

// persistVectors 事务批量写入向量记录到 file_vectors 表。
// 利用数据库层 (file_id, number) 唯一索引（WHERE deleted_at IS NULL）保证幂等。
func persistVectors(fileID uint64, items []EmbedVectorItem) error {
	if len(items) == 0 {
		return nil
	}

	db := infra.GetDB()

	return db.Transaction(func(tx *gorm.DB) error {
		// 二次确认：事务内再次检查幂等（防止并发场景下重复写入）
		var existingCount int64
		if err := tx.Model(&model.FileVector{}).
			Where("file_id = ?", fileID).
			Count(&existingCount).Error; err != nil {
			return fmt.Errorf("transactional idempotency check failed: %w", err)
		}
		if existingCount > 0 {
			log.Printf("[file-embed] File %d vectors already exist (transactional check), skipping", fileID)
			return nil
		}

		records := make([]model.FileVector, 0, len(items))
		for _, item := range items {
			records = append(records, model.FileVector{
				FileID: fileID,
				Number: item.Number,
				Vector: item.Vector,
			})
		}

		// ON CONFLICT DO NOTHING：若并发下另一事务已写入相同 (file_id, number)，
		// 则静默跳过，不报错。
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&records)
		if result.Error != nil {
			// 忽略唯一约束冲突之外的错误
			errStr := result.Error.Error()
			if strings.Contains(errStr, "duplicate") || strings.Contains(errStr, "23505") {
				log.Printf("[file-embed] Ignoring duplicate key for file_id=%d", fileID)
				return nil
			}
			return fmt.Errorf("insert file_vectors: %w", result.Error)
		}

		log.Printf("[file-embed] Inserted %d / %d vectors for file_id=%d",
			result.RowsAffected, len(records), fileID)
		return nil
	})
}

package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"github.com/lojes7/inquire/internal/model"
	"github.com/lojes7/inquire/internal/ws"
	"github.com/lojes7/inquire/pkg/infra"
	"github.com/lojes7/inquire/pkg/secure"
	"github.com/lojes7/inquire/pkg/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const fileHashUniqueIndexName = "idx_files_hash_value"

var errFileHashConflict = errors.New("file hash unique conflict")

// notifyConversationUsers 通知会话中的所有用户
func notifyConversationUsers(conversationID uint64, msgType string, data any) {
	var userIDs []uint64
	err := infra.GetDB().Model(&model.ConversationUser{}).
		Where("conversation_id = ?", conversationID).
		Pluck("user_id", &userIDs).Error
	if err != nil {
		log.Printf("获取会话用户失败: %v\n", err)
		return
	}

	payload := map[string]any{
		"type": msgType,
		"data": data,
	}
	msgBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("json marshal failed: %v\n", err)
		return
	}

	for _, uid := range userIDs {
		ws.GetHub().SendToUser(uid, msgBytes)
	}
}

// sendMessageAuth 验证用户是否有权限在该会话中发送消息
func sendMessageAuth(userID, conversationID uint64) error {
	// 检查 conversation_users 表中是否存在该用户和会话，且未被禁言
	var cu model.ConversationUser
	db := infra.GetDB()
	err := db.Where("user_id = ? AND conversation_id = ?", userID, conversationID).
		First(&cu).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return secure.Wrap(403, "无权限在该会话中发送消息", errors.New("forbidden"))
		}
		log.Println(err)
		return secure.Wrap(500, "验证权限失败", err)
	}

	if cu.IsBanned {
		return secure.Wrap(403, "您已被禁言/拉黑，无法发送消息", errors.New("banned"))
	}

	return nil
}

// createSystemMessage 在一个会话中创建一个系统级消息
// newID用户指定该系统消息的ID
func createSystemMessage(tx *gorm.DB, content string, conversationID, newID uint64) error {
	newMsg := model.Message{
		SenderID:       0,
		ConversationID: conversationID,
		MyModel: model.MyModel{
			ID: newID,
		},
		Status:  model.SYSTEM,
		Content: content,
	}

	res := tx.Create(&newMsg)
	if res.Error != nil {
		log.Println(res.Error)
		return secure.Wrap(500, "创建系统消息失败", res.Error)
	}

	return nil
}

// updateUnreadCount 给当前会话中除开当前sender的所有人的unread_count加一
func updateUnreadCount(tx *gorm.DB, senderID, conversationID uint64) error {
	res := tx.Model(&model.ConversationUser{}).
		Where("user_id != ? AND conversation_id = ?",
			senderID, conversationID).
		UpdateColumn("unread_count", gorm.Expr("unread_count + ?", 1))
	if res.Error != nil {
		log.Println(res.Error)
		return secure.Wrap(500, "更新未读数失败", res.Error)
	}

	if res.RowsAffected == 0 {
		log.Println("更新unread count字段影响了0行表")
		return secure.Wrap(500, "更新未读数失败", errors.New("rows affected 0"))
	}
	return nil
}

// updateLastMessageID 更新当前会话的last_message_id
func updateLastMessageID(tx *gorm.DB, conversationID, msgID uint64) error {
	res := tx.Model(&model.ConversationUser{}).
		Where("conversation_id = ?", conversationID).
		Update("last_message_id", msgID)
	if res.Error != nil {
		log.Println(res.Error)
		return secure.Wrap(500, "更新最新消息失败", res.Error)
	}
	if res.RowsAffected == 0 {
		log.Println("更新last msg id字段影响了0行表")
		return secure.Wrap(500, "更新最新消息失败", errors.New("rows affected 0"))
	}
	return nil
}

func SendText(senderID, conversationID uint64, content string) (uint64, error) {
	err := sendMessageAuth(senderID, conversationID)
	if err != nil {
		return 0, err
	}

	newID := utils.NewUniqueID()
	newMsg := model.Message{
		SenderID:       senderID,
		ConversationID: conversationID,
		Status:         model.TEXT,
		MyModel: model.MyModel{
			ID: newID,
		},
		Content: content,
	}

	db := infra.GetDB()
	err = db.Transaction(func(tx *gorm.DB) error {
		res := tx.Create(&newMsg)
		if res.Error != nil {
			log.Println(res.Error)
			return secure.Wrap(500, "发送消息失败", res.Error)
		}

		err := updateLastMessageID(tx, conversationID, newID)
		if err != nil {
			return err
		}

		err = updateUnreadCount(tx, senderID, conversationID)
		if err != nil {
			return err
		}

		return nil
	})

	if err == nil {
		// 发送 websocket 通知
		notifyConversationUsers(conversationID, "new_message", map[string]any{
			"message_id":      newID,
			"conversation_id": conversationID,
			"sender_id":       senderID,
			"content":         content,
			"status":          model.TEXT, // 0
			"updated_at":      time.Now(),
		})
	}
	return newID, err
}

func SendFile(ctx context.Context, senderID, conversationID uint64, file *multipart.FileHeader) (*model.SendFileResp, error) {
	// 先做会话权限校验，未通过直接返回
	err := sendMessageAuth(senderID, conversationID)
	if err != nil {
		return nil, err
	}

	// 新消息的 id
	newMsgID := utils.NewUniqueID()

	// 先计算文件内容哈希，用于跨文件名去重。
	// 只要二进制内容相同，即使文件名不同也会命中同一 hash。
	hashValue, err := computeFileHash(file)
	if err != nil {
		return nil, err
	}

	resp := &model.SendFileResp{MessageID: newMsgID}
	var savedFilePath string

	db := infra.GetDB()
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1) 先按 hash 查询 files 表，命中则直接复用，避免重复落盘与重复入库。
		fileRecord, existed, err := checkFileByHash(tx, hashValue)
		if err != nil {
			return err
		}
		if existed {
			if strings.TrimSpace(fileRecord.FileURL) == "" {
				return secure.Wrap(500, "文件记录异常", errors.New("existing file url is empty"))
			}

			// 命中重复文件时，仍然创建消息，但复用已有 file_id。
			newMsg := model.Message{
				MyModel: model.MyModel{
					ID: newMsgID,
				},
				SenderID:       senderID,
				ConversationID: conversationID,
				Status:         model.FILE,
				FileID:         fileRecord.ID,
			}

			res := tx.Create(&newMsg)
			if res.Error != nil {
				log.Println(res.Error)
				return secure.Wrap(500, "发送文件消息失败", res.Error)
			}

			err = createUserFileRelations(tx, conversationID, fileRecord.ID)
			if err != nil {
				return err
			}

			err = updateLastMessageID(tx, conversationID, newMsgID)
			if err != nil {
				return err
			}

			err = updateUnreadCount(tx, senderID, conversationID)
			if err != nil {
				return err
			}

			resp.FileName = fileRecord.FileName
			resp.FileSize = fileRecord.FileSize
			resp.FileType = fileRecord.FileType
			return nil
		}

		// 2) 未命中去重时：先落盘，再写 files 表（完整信息一次写入）。
		//    这样可避免 files 出现 file_url 为空的占位中间态。
		savedFileInfo, saveErr := SaveFileIntoServer(file)
		if saveErr != nil {
			return saveErr
		}
		savedFilePath = savedFileInfo.FilePath

		fileRecord, err = createFileRecord(tx, hashValue, savedFileInfo)
		if err != nil {
			if errors.Is(err, errFileHashConflict) {
				// 并发补偿：若另一个事务已先写入同 hash，当前刚落盘文件应清理，随后复用已有记录。
				existingFile, existedErr := getExistingFileByHash(tx, hashValue)
				if existedErr != nil {
					return existedErr
				}
				if existingFile == nil {
					return secure.Wrap(500, "文件重复校验失败", errors.New("hash conflict but file not found"))
				}

				if removeErr := os.Remove(savedFilePath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
					log.Printf("并发冲突后清理重复落盘文件失败: %v", removeErr)
				}
				savedFilePath = ""
				fileRecord = existingFile
			} else {
				return err
			}
		}

		if strings.TrimSpace(fileRecord.FileURL) == "" {
			return secure.Wrap(500, "文件记录异常", errors.New("file url is empty"))
		}

		// 3) 无论是否重复文件，都创建一条消息。
		//    当命中重复时，消息会复用已有 file_id，
		//    从而保证“聊天消息行为”不受去重影响。
		newMsg := model.Message{
			MyModel: model.MyModel{
				ID: newMsgID,
			},
			SenderID:       senderID,
			ConversationID: conversationID,
			Status:         model.FILE,
			FileID:         fileRecord.ID,
		}

		res := tx.Create(&newMsg)
		if res.Error != nil {
			log.Println(res.Error)
			return secure.Wrap(500, "发送文件消息失败", res.Error)
		}

		// 4) 维护关联关系与会话状态。
		//    user_files 使用数据库唯一索引去重（ON CONFLICT DO NOTHING），
		//    避免应用层重复判断与并发窗口问题。
		err = createUserFileRelations(tx, conversationID, fileRecord.ID)
		if err != nil {
			return err
		}

		err = updateLastMessageID(tx, conversationID, newMsgID)
		if err != nil {
			return err
		}

		err = updateUnreadCount(tx, senderID, conversationID)
		if err != nil {
			return err
		}

		resp.FileName = fileRecord.FileName
		resp.FileSize = fileRecord.FileSize
		resp.FileType = fileRecord.FileType
		return nil
	})

	if err != nil {
		// 只有“本次请求新落盘了文件”才需要做磁盘回滚；
		// 若本次复用了已有 file 记录，则 savedFilePath 为空，不会误删历史文件。
		if strings.TrimSpace(savedFilePath) != "" {
			if removeErr := os.Remove(savedFilePath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
				log.Printf("数据库写入失败后清理文件失败: %v", removeErr)
			}
		}
		return nil, err
	}

	// 发送 websocket 通知
	// resp is: MessageID, FileName, FileSize, FileType
	notifyConversationUsers(conversationID, "new_message", map[string]any{
		"message_id":      newMsgID,
		"conversation_id": conversationID,
		"sender_id":       senderID,
		"content":         resp,
		"status":          model.FILE, // 3
		"updated_at":      time.Now(),
	})

	return resp, nil
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

func DownloadFile(userID, messageID uint64) (string, error) {
	db := infra.GetDB()

	// 一次查询完成：消息存在 + 用户在对话中 + 文件存在
	var file model.File
	err := db.Model(&model.File{}).
		Select("files.file_url").
		Joins("JOIN messages m ON m.file_id = files.id").
		Joins("JOIN conversation_users cu ON cu.conversation_id = m.conversation_id").
		Where("files.message_id = ? AND cu.user_id = ? AND m.status = ?",
			messageID, userID, model.FILE).
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

func RecallMessage(userID, msgID uint64) (uint64, error) {
	db := infra.GetDB()
	var temp model.Message
	err := db.Model(&model.Message{}).
		Select("sender_id, conversation_id").
		Where("id = ?", msgID).
		First(&temp).
		Error
	if err != nil {
		log.Println(err)
		return 0, secure.Wrap(500, "查询消息失败", err)
	}

	senderID := temp.SenderID
	conversationID := temp.ConversationID
	if senderID != userID {
		return 0, secure.Wrap(403, "不能撤回不是自己发的消息", errors.New("forbidden"))
	}

	newID := utils.NewUniqueID()
	var newContent string // Declare variable to capture content inside transaction

	err = db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.Message{}).
			Where("id = ?", msgID).
			Update("status", model.RECALLED)
		if res.Error != nil {
			log.Println(res.Error)
			return secure.Wrap(500, "撤回消息失败", res.Error)
		}
		if res.RowsAffected == 0 {
			log.Println("撤回消息操作影响了0行表")
			return secure.Wrap(500, "撤回消息失败", errors.New("rows affected 0"))
		}

		var senderName string
		err = tx.Model(&model.User{}).
			Where("id = ?", senderID).
			Pluck("name", &senderName).Error
		if err != nil {
			log.Println(err)
			return secure.Wrap(500, "获取用户名称失败", err)
		}
		newContent = senderName + "撤回了一条消息" // Assign to captured variable
		err = createSystemMessage(tx, newContent, conversationID, newID)
		if err != nil {
			return err
		}

		err = updateLastMessageID(tx, conversationID, newID)
		if err != nil {
			return err
		}
		return nil
	})

	if err == nil {
		// 发送 websocket 通知
		// 1. Tell clients to update the old message to RECALLED status
		notifyConversationUsers(conversationID, "recall_message", map[string]any{
			"recalled_message_id": msgID,
			"system_message_id":   newID,
			"conversation_id":     conversationID,
			"content":             newContent,
			"updated_at":          time.Now(),
		})
	}

	return newID, err
}

// DeleteMessage 删除消息 逻辑是：
func DeleteMessage(userID, messageID uint64) error {
	db := infra.GetDB()
	var msg model.Message

	err := db.Model(&model.Message{}).
		Select("conversation_id").
		Where("id = ?", messageID).
		First(&msg).
		Error

	if err != nil {
		log.Println(err)
		return secure.Wrap(500, "没有找到该条消息", err)
	}
	conversationID := msg.ConversationID

	err = db.Transaction(func(tx *gorm.DB) error {
		res := tx.Where("user_id = ? AND message_id = ?", userID, messageID).
			Delete(&model.MessageUser{})

		if res.Error != nil {
			log.Println(res.Error)
			return secure.Wrap(500, "删除消息失败", res.Error)
		}
		if res.RowsAffected == 0 {
			log.Println("删除消息操作影响了0行表，自动创建删除记录")
			newID := utils.NewUniqueID()
			newMU := model.MessageUser{
				MyModel:   model.MyModel{ID: newID},
				UserID:    userID,
				MessageID: messageID,
				IsDeleted: true,
			}
			if err := tx.Create(&newMU).Error; err != nil {
				log.Println("创建已删除的MessageUser失败:", err)
				return secure.Wrap(500, "删除消息失败", err)
			}
		}

		var lastID uint64
		sql := `SELECT m.id 
			FROM messages m 
			LEFT JOIN message_users mu 
			    ON mu.message_id = m.id AND mu.user_id = ? AND mu.deleted_at IS NULL
			WHERE m.status != ? AND mu.is_deleted = false AND m.deleted_at IS NULL
			ORDER BY m.created_at DESC 
			LIMIT 1`
		res = tx.Raw(sql, userID, model.RECALLED).Scan(&lastID)
		if res.Error != nil {
			log.Println(res.Error)
			return secure.Wrap(500, "更新最新消息失败", res.Error)
		}
		if res.RowsAffected == 0 {
			log.Println("删除消息 更新最后消息id时没有查到id")
			return secure.Wrap(500, "更新最新消息失败", errors.New("rows affected 0"))
		}

		err = updateLastMessageID(tx, conversationID, lastID)
		if err != nil {
			return err
		}

		return nil
	})

	if err == nil {
		// Only notify the user who performed the delete
		payload := map[string]any{
			"type": "delete_message",
			"data": map[string]any{
				"message_id":      messageID,
				"conversation_id": conversationID,
			},
		}
		msgBytes, _ := json.Marshal(payload)
		ws.GetHub().SendToUser(userID, msgBytes)
	}

	return err
}

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/lojes7/inquire/internal/model"
	"github.com/lojes7/inquire/internal/ws"
	"github.com/lojes7/inquire/pkg/infra"
	"github.com/lojes7/inquire/pkg/secure"
	"github.com/lojes7/inquire/pkg/utils"
	"gorm.io/gorm"
)

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
	// 检查 conversation_users 表中是否存在该用户和会话
	var cnt int64
	db := infra.GetDB()
	err := db.Model(&model.ConversationUser{}).
		Where("user_id = ? AND conversation_id = ?", userID, conversationID).
		Count(&cnt).Error
	if err != nil {
		log.Println(err)
		return secure.Wrap(500, "验证权限失败", err)
	}
	if cnt == 0 {
		return secure.Wrap(403, "无权限在该会话中发送消息", errors.New("forbidden"))
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
		Status: model.SYSTEM,
	}
	newText := model.Text{
		Text:      content,
		MessageID: newID,
	}
	res := tx.Create(&newMsg)
	if res.Error != nil {
		log.Println(res.Error)
		return secure.Wrap(500, "创建系统消息失败", res.Error)
	}

	res = tx.Create(&newText)
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
	}
	newText := model.Text{
		Text:      content,
		MessageID: newID,
	}
	db := infra.GetDB()
	err = db.Transaction(func(tx *gorm.DB) error {
		res := tx.Create(&newMsg)
		if res.Error != nil {
			log.Println(res.Error)
			return secure.Wrap(500, "发送消息失败", res.Error)
		}

		res = tx.Create(&newText)
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
	err := sendMessageAuth(senderID, conversationID)
	if err != nil {
		return nil, err
	}

	newID := utils.NewUniqueID()

	// 获取 uploads 目录
	uploadDir := infra.GetFilePath()

	// 生成文件路径，使用newID作为文件名，保持原扩展名
	ext := filepath.Ext(file.Filename)
	fileName := strings.TrimSuffix(file.Filename, ext) // 原文件名
	filePath := filepath.Join(uploadDir, fmt.Sprintf("%d%s", newID, ext))

	// 保存文件
	if err := saveFile(file, filePath); err != nil {
		log.Println(err)
		return nil, secure.Wrap(500, "保存文件失败", err)
	}

	// 获取文件信息
	fileSize := file.Size
	fileType := getFileType(filePath)
	fileContent, contentVector, err := buildFileIndex(ctx, filePath, fileName, fileType)
	if err != nil {
		return nil, err
	}

	newMsg := model.Message{
		SenderID:       senderID,
		ConversationID: conversationID,
		Status:         model.FILE,
		MyModel: model.MyModel{
			ID: newID,
		},
	}

	// 新文件
	newFile := model.File{
		FileName:      fileName,
		FileType:      fileType,
		FileURL:       filePath,
		FileSize:      fileSize,
		MessageID:     newID,
		FileContent:   fileContent,
		ContentVector: contentVector,
	}

	// 保存到数据库
	db := infra.GetDB()
	resp := &model.SendFileResp{
		MessageID: newID,
		FileName:  fileName,
		FileSize:  fileSize,
		FileType:  fileType,
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		res := tx.Create(&newMsg)
		if res.Error != nil {
			log.Println(res.Error)
			return secure.Wrap(500, "发送文件消息失败", res.Error)
		}

		res = tx.Create(&newFile)
		if res.Error != nil {
			log.Println(res.Error)
			return secure.Wrap(500, "发送文件消息失败", res.Error)
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
		// actually frontend might need different structure
		// resp is: MessageID, FileName, FileSize, FileType
		notifyConversationUsers(conversationID, "new_message", map[string]any{
			"message_id":      newID,
			"conversation_id": conversationID,
			"sender_id":       senderID,
			"content":         resp,       // Sending the file response object as content
			"status":          model.FILE, // 3
			"updated_at":      time.Now(),
		})
	}

	return resp, err
}

func DownloadFile(userID, messageID uint64) (string, error) {
	db := infra.GetDB()

	// 一次查询完成：消息存在 + 用户在对话中 + 文件存在
	var file model.File
	err := db.Model(&model.File{}).
		Select("files.file_url").
		Joins("JOIN messages m ON m.id = files.message_id").
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
		return secure.Wrap(500, "查询消息失败", err)
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
			log.Println("删除消息操作影响了0行表")
			return secure.Wrap(500, "删除消息失败", errors.New("rows affected 0"))
		}

		var lastID uint64
		sql := `SELECT m.id FROM messages m 
			LEFT JOIN message_users mu ON mu.message_id = m.id AND mu.user_id = ?
			WHERE m.status != ? AND mu.is_deleted = false
			ORDER BY m.created_at DESC 
			LIMIT 1`
		res = tx.Raw(sql, userID, model.RECALLED).Scan(&lastID)
		if res.Error != nil {
			log.Println(res.Error)
			return secure.Wrap(500, "更新最新消息失败", res.Error)
		}
		if res.RowsAffected == 0 {
			log.Println("删除消息更新最后消息id 时没有查到id")
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

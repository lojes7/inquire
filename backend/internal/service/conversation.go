package service

import (
	"errors"
	"log"
	"vvechat/internal/model"
	"vvechat/pkg/infra"
	"vvechat/pkg/utils"

	"gorm.io/gorm"
)

// CreatePrivateConversation 新建私聊
func CreatePrivateConversation(userID, friendID uint64) error {
	db := infra.GetDB()
	var converse model.Conversation

	conversationID := utils.NewUniqueID()
	converse.ID = conversationID
	converse.Type = model.PRIVATE

	return db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.Conversation{}).
			Create(&converse)
		if res.Error != nil {
			log.Println(res.Error)
			if errors.Is(res.Error, gorm.ErrDuplicatedKey) {
				return errors.New("错误：重复创建 conversation")
			}
			return errors.New("服务器错误")
		}
		if res.RowsAffected == 0 {
			log.Println("创建conversation操作影响了0行表")
			return errors.New("服务器错误")
		}

		res = tx.Exec(`INSERT INTO conversation_users 
			(id, user_id, conversation_id, remark)
			SELECT ?, ?, ?, f.friend_remark
			FROM friendships f 
			WHERE f.friend_id = ?`,
			utils.NewUniqueID(), userID, conversationID, friendID)
		if res.Error != nil {
			log.Println(res.Error)
			return errors.New("服务器错误")
		}
		if res.RowsAffected == 0 {
			log.Println("创建conversation_users操作影响了0行表")
			return errors.New("服务器错误")
		}

		return nil
	})
}

// ChatHistoryList 加载聊天记录
func ChatHistoryList(userID, conversationID uint64) ([]model.ChatHistoryResp, error) {
	db := infra.GetDB()

	resp := make([]model.ChatHistoryResp, 0)

	sql := `SELECT m.id AS message_id, 
       		m.sender_id, 
       		u.name AS sender_name,
			m.status, 
			m.updated_at,
			CASE m.status
			WHEN ? OR ? THEN t.text
			WHEN ? THEN json_build_object(
               'file_name', f.file_name,
               'file_url', f.file_url,
               'file_size', f.file_size,
               'file_type', f.file_type
           )
			ELSE ''
			END AS content
			FROM messages m
			LEFT JOIN users u ON u.id = m.sender_id
			LEFT JOIN message_users mu ON mu.message_id = m.id AND mu.user_id = ?
			LEFT JOIN texts t ON t.message_id = m.id
			LEFT JOIN files f ON f.message_id = m.id
			WHERE mu.is_deleted = false AND m.conversation_id = ? AND m.status != ?
			ORDER BY m.updated_at DESC`

	res := db.Raw(sql, model.TEXT,
		model.SYSTEM,
		model.FILE,
		userID,
		conversationID,
		model.RECALLED).
		Scan(&resp)

	if res.Error != nil {
		log.Println(res.Error)
		return nil, errors.New("服务器错误")
	}
	return resp, nil
}

// ConversationList 会话列表
func ConversationList(userID uint64) ([]model.ConversationListResp, error) {
	db := infra.GetDB()
	resp := make([]model.ConversationListResp, 0)

	sql := `SELECT cu.remark, 
       	cu.conversation_id,
       	cu.unread_count,
       	CASE m.status
		WHEN ? OR ? THEN t.text
		WHEN ? THEN f.file_name
		ELSE ''
		END AS content
		FROM conversation_users cu 
		LEFT JOIN messages m ON m.id = cu.last_message_id
		LEFT JOIN files f ON f.message_id = m.id
		LEFT JOIN texts t ON t.message_id = m.id
		WHERE cu.user_id = ? AND cu.is_deleted = false
		ORDER BY cu.is_pinned DESC, cu.updated_at DESC `

	res := db.Raw(sql, model.TEXT,
		model.SYSTEM,
		model.FILE,
		userID).
		Scan(&resp)

	if res.Error != nil {
		log.Println(res.Error)
		return nil, errors.New("服务器错误")
	}

	return resp, nil
}

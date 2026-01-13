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

// EnterConversation 进入聊天窗口
func EnterConversation(userID, conversationID uint64) ([]model.EnterConversationResp, error) {
	db := infra.GetDB()

	resp := make([]model.EnterConversationResp, 0)

	res := db.Raw(`SELECT m.content, m.id, m.status, u.name
		FROM messages m 
		JOIN users u ON u.id = m.sender_id
		LEFT JOIN message_users mu ON mu.user_id = ? AND mu.message_id = m.id
		WHERE m.conversation_id = ? AND  m.status != 1 AND mu.is_deleted = false
		ORDER BY m.updated_at DESC `, userID, conversationID).Find(&resp)

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

	sql := `SELECT m.content, cu.remark, cu.conversation_id
		FROM conversation_users cu 
		LEFT JOIN messages m ON m.id = cu.last_message_id
		WHERE cu.user_id = ?
		ORDER BY cu.is_pinned DESC, m.updated_at DESC `
	res := db.Raw(sql, userID).Find(&resp)

	if res.Error != nil {
		log.Println(res.Error)
		return nil, errors.New("服务器错误")
	}

	return resp, nil
}

// updateUnreadCount 给当前会话中除开当前sender的所有人的unread_count加一
func updateUnreadCount(tx *gorm.DB, senderID, conversationID uint64) error {
	res := tx.Model(&model.ConversationUser{}).
		Where("user_id != ? AND conversation_id = ?",
			senderID, conversationID).
		UpdateColumn("unread_count", gorm.Expr("unread_count + ?", 1))
	if res.Error != nil {
		log.Println(res.Error)
		return res.Error
	}
	if res.RowsAffected == 0 {
		log.Println("更新unread count字段影响了0行表")
		return errors.New("更新unread count失败")
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
		return res.Error
	}
	if res.RowsAffected == 0 {
		log.Println("更新last msg id字段影响了0行表")
		return errors.New("更新last msg id失败")
	}
	return nil
}

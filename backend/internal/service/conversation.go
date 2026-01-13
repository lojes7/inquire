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
func CreatePrivateConversation(userID, conversationID uint64) error {
	db := infra.GetDB()
	var converse model.Conversation

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
			utils.NewUniqueID(), userID, conversationID, conversationID)
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

func EnterConversation(conversationID uint64) ([]model.EnterConversationResp, error) {
	db := infra.GetDB()

	resp := make([]model.EnterConversationResp, 0)

	res := db.Raw(`SELECT m.content, m.id, m.status, u.name
		FROM messages m 
		JOIN users u ON u.id = m.sender_id
		LEFT JOIN message_users mu ON mu.user_id = u.id AND mu.message_id = m.id
		WHERE m.conversation_id = ? AND  m.status != 1 AND mu.is_deleted = false
		ORDER BY m.updated_at DESC `, conversationID).Find(&resp)

	if res.Error != nil {
		log.Println(res.Error)
		return nil, errors.New("服务器错误")
	}
	return resp, nil
}

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

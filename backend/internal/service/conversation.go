package service

import (
	"errors"
	"log"

	"github.com/lojes7/inquire/internal/model"
	"github.com/lojes7/inquire/pkg/infra"
	"github.com/lojes7/inquire/pkg/utils"
	"gorm.io/gorm"
)

// StartPrivateConversation 发起私聊
// 会调用getPrivateConversationID以获取会话ID
// 最后返回会话ID
func StartPrivateConversation(userID, friendID uint64) (uint64, error) {
	// 找到 A 和 B 共同的 conversation_id
	conversationID, err := getPrivateConversationID(userID, friendID)
	if err != nil {
		return 0, err
	}

	return conversationID, nil
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
			WHERE m.conversation_id = ? AND m.status != ? AND mu.is_deleted = false
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
		WHERE cu.user_id = ? 
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

// getPrivateConversationID 获取两用户之间的私聊会话ID
// 两用户是好友关系才能正常工作，若不存在会话则创建新会话
func getPrivateConversationID(userID, friendID uint64) (uint64, error) {
	ok, err := isFriend(userID, friendID)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, errors.New("两用户不是好友关系")
	}

	db := infra.GetDB()
	var conversationID uint64

	res := db.Raw(`
    SELECT cu1.conversation_id
    FROM conversation_users cu1
    JOIN conversation_users cu2 ON cu1.conversation_id = cu2.conversation_id
    WHERE (cu1.user_id = ? AND cu2.user_id = ?)
       OR (cu1.user_id = ? AND cu2.user_id = ?)`,
		userID, friendID,
		friendID, userID).Scan(&conversationID)

	if res.Error != nil {
		// 数据库出错
		log.Println(res.Error)
		return 0, errors.New("服务器错误")
	}

	if res.RowsAffected == 0 {
		return createPrivateConversation(userID, friendID)
	}

	return conversationID, nil
}

// createPrivateConversation 创建两用户之间的私聊会话
// 使用前需要严格确保两用户之前不存在会话，且需要确保两用户是好友关系
func createPrivateConversation(userID, friendID uint64) (uint64, error) {
	db := infra.GetDB()
	newID := utils.NewUniqueID()

	err := db.Transaction(func(tx *gorm.DB) error {
		// 查询用户对好友的备注
		userToFriendRemark, err := getFriendRemark(tx, userID, friendID)
		if err != nil {
			return err
		}

		// 查询好友对用户的备注
		friendToUserRemark, err := getFriendRemark(tx, friendID, userID)
		if err != nil {
			return err
		}

		// 先创建出一个新 conversation
		c := model.Conversation{}
		c.ID = newID
		c.Type = model.PRIVATE

		res := tx.Create(&c)
		if res.Error != nil {
			log.Println(res.Error)
			return errors.New("创建会话失败")
		}

		// 在 conversation_users 表中创建出新 conversation，填入查到的备注
		if err := createConversationUser(tx, userID, newID, userToFriendRemark); err != nil {
			return err
		}

		if err := createConversationUser(tx, friendID, newID, friendToUserRemark); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return newID, nil
}

// createConversationUser 用于创建出一个 conversation_users 表的字段
func createConversationUser(tx *gorm.DB, userID, conversationID uint64, remark string) error {
	if tx == nil {
		tx = infra.GetDB()
	}

	cu := model.ConversationUser{
		UserID:         userID,
		ConversationID: conversationID,
		Remark:         remark,
	}

	res := tx.Create(&cu)
	if res.Error != nil {
		log.Println(res.Error)
		return errors.New("服务器错误")
	}

	return nil
}

func deleteConversationUser(tx *gorm.DB, userID, conversationID uint64) error {
	if tx == nil {
		tx = infra.GetDB()
	}

	res := tx.Where("user_id = ? AND conversation_id = ?", userID, conversationID).
		Delete(&model.ConversationUser{})

	if res.Error != nil {
		log.Println(res.Error)
		return errors.New("服务器错误")
	}
	if res.RowsAffected == 0 {
		log.Println("删除 conversation_users 操作影响了0行表")
	}

	return nil
}

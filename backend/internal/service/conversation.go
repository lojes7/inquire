package service

import (
	"errors"
	"log"

	"github.com/lojes7/inquire/internal/model"
	"github.com/lojes7/inquire/pkg/infra"
	"github.com/lojes7/inquire/pkg/secure"
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
			CASE 
			WHEN m.status IN (?, ?) THEN json_build_object('text', t.text)
			WHEN m.status = ? THEN json_build_object(
               'file_name', f.file_name,
               'file_url', f.file_url,
               'file_size', f.file_size,
               'file_type', f.file_type
           )
			ELSE '{}'::json
			END AS content
			FROM messages m
			LEFT JOIN users u ON u.id = m.sender_id AND u.deleted_at IS NULL
			LEFT JOIN message_users mu ON mu.message_id = m.id AND mu.user_id = ? AND mu.deleted_at IS NULL
			LEFT JOIN texts t ON t.message_id = m.id AND t.deleted_at IS NULL
			LEFT JOIN files f ON f.message_id = m.id AND f.deleted_at IS NULL
			WHERE m.conversation_id = ? AND m.status != ? AND m.deleted_at IS NULL AND (mu.is_deleted = false OR mu.is_deleted IS NULL)
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
		return nil, secure.Wrap(500, "加载聊天记录失败", res.Error)
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
       	CASE 
  			WHEN m.status IN (?, ?) THEN t.text
  			WHEN m.status = ? THEN f.file_name
  			ELSE ''
		END AS content
		FROM conversation_users cu 
		LEFT JOIN messages m ON m.id = cu.last_message_id AND m.deleted_at IS NULL
		LEFT JOIN files f ON f.message_id = m.id AND f.deleted_at IS NULL
		LEFT JOIN texts t ON t.message_id = m.id AND t.deleted_at IS NULL
		WHERE cu.user_id = ? AND cu.deleted_at IS NULL
		ORDER BY cu.is_pinned DESC, cu.updated_at DESC `

	res := db.Raw(sql, model.TEXT,
		model.SYSTEM,
		model.FILE,
		userID).
		Scan(&resp)

	if res.Error != nil {
		log.Println(res.Error)
		return nil, secure.Wrap(500, "加载会话列表失败", res.Error)
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
		return 0, secure.Wrap(400, "两用户不是好友关系", errors.New("not friend"))
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
		return 0, secure.Wrap(500, "查询会话失败", res.Error)
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
			return secure.Wrap(500, "创建会话失败", res.Error)
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

	// 检查用户是否存在
	var userCount int64
	if err := tx.Model(&model.User{}).
		Where("id = ?", userID).
		Count(&userCount).Error; err != nil {
		return secure.Wrap(500, "检查用户存在状态失败", err)
	}
	if userCount == 0 {
		return secure.Wrap(404, "用户不存在", errors.New("user not found"))
	}

	// 检查会话是否存在
	var convCount int64
	if err := tx.Model(&model.Conversation{}).
		Where("id = ?", conversationID).
		Count(&convCount).Error; err != nil {
		return secure.Wrap(500, "检查会话存在状态失败", err)
	}
	if convCount == 0 {
		return secure.Wrap(404, "会话不存在", errors.New("conversation not found"))
	}

	cu := model.ConversationUser{
		UserID:         userID,
		ConversationID: conversationID,
		Remark:         remark,
	}

	res := tx.Create(&cu)
	if res.Error != nil {
		log.Println(res.Error)
		return secure.Wrap(500, "创建会话成员失败", res.Error)
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
		return secure.Wrap(500, "删除会话成员失败", res.Error)
	}
	if res.RowsAffected == 0 {
		log.Println("删除 conversation_users 操作影响了0行表")
	}

	return nil
}

// CreateGroupConversation 创建群聊
func CreateGroupConversation(ownerID uint64, groupName string, memberIDs []uint64) (uint64, error) {
	db := infra.GetDB()
	newID := utils.NewUniqueID()

	err := db.Transaction(func(tx *gorm.DB) error {
		// 创建 Conversation
		c := model.Conversation{
			Type:      model.GROUP,
			OwnerID:   ownerID,
			GroupName: groupName,
		}
		c.ID = newID

		if err := tx.Create(&c).Error; err != nil {
			return secure.Wrap(500, "创建群聊失败", err)
		}

		// 创建成员
		for _, uid := range memberIDs {
			// 将群名称作为所有成员的备注
			if err := createConversationUser(tx, uid, newID, groupName); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return 0, err
	}
	return newID, nil
}

// BanUser 禁言/拉黑用户
// 私聊：A 拉黑 B -> B 的 is_banned = true
// 群聊：群主拉黑 B -> B 的 is_banned = true
func BanUser(operatorID, conversationID, targetUserID uint64) error {
	db := infra.GetDB()
	var conversation model.Conversation
	if err := db.First(&conversation, conversationID).Error; err != nil {
		return secure.Wrap(404, "会话不存在", err)
	}

	if conversation.Type == model.PRIVATE {
		// 私聊：验证 operatorID 是否在对话中
		// 获取双方UID
		var userIDs []uint64
		if err := db.Model(&model.ConversationUser{}).Where("conversation_id = ?", conversationID).Pluck("user_id", &userIDs).Error; err != nil {
			return secure.Wrap(500, "获取会话成员失败", err)
		}

		isParticipant := false
		for _, uid := range userIDs {
			if uid == operatorID {
				isParticipant = true
				break
			}
		}
		if !isParticipant {
			return secure.Wrap(403, "无权操作", errors.New("not participant"))
		}

		// 目标必须是对方
		if operatorID == targetUserID {
			return secure.Wrap(400, "不能拉黑自己", errors.New("cannot ban self"))
		}

		// 验证 targetUserID 是否也在对话中
		targetIn := false
		for _, uid := range userIDs {
			if uid == targetUserID {
				targetIn = true
				break
			}
		}
		if !targetIn {
			return secure.Wrap(400, "目标不在会话中", errors.New("target not in conversation"))
		}

	} else if conversation.Type == model.GROUP {
		// 群聊：验证 operatorID 是否是群主
		if conversation.OwnerID != operatorID {
			return secure.Wrap(403, "非群主无权禁言", errors.New("not owner"))
		}
	}

	// 执行禁言
	res := db.Model(&model.ConversationUser{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, targetUserID).
		Update("is_banned", true)

	if res.Error != nil {
		return secure.Wrap(500, "禁言失败", res.Error)
	}
	if res.RowsAffected == 0 {
		return secure.Wrap(404, "目标用户不在会话中", errors.New("user not found in conversation"))
	}

	return nil
}

// KickUser 踢出群聊
// 仅群主可用
func KickUser(operatorID, conversationID, targetUserID uint64) error {
	db := infra.GetDB()
	var conversation model.Conversation
	if err := db.First(&conversation, conversationID).Error; err != nil {
		return secure.Wrap(404, "会话不存在", err)
	}

	if conversation.Type != model.GROUP {
		return secure.Wrap(400, "非群聊不能踢人", errors.New("not group chat"))
	}

	if conversation.OwnerID != operatorID {
		return secure.Wrap(403, "非群主无权踢人", errors.New("not owner"))
	}

	if operatorID == targetUserID {
		return secure.Wrap(400, "不能踢自己", errors.New("cannot kick self"))
	}

	return deleteConversationUser(db, targetUserID, conversationID)
}

// LeaveGroup 退出群聊
// 如果是群主退出，则解散群聊
func LeaveGroup(userID, conversationID uint64) error {
	db := infra.GetDB()
	var conversation model.Conversation
	if err := db.First(&conversation, conversationID).Error; err != nil {
		return secure.Wrap(404, "会话不存在", err)
	}

	if conversation.Type == model.PRIVATE {
		return secure.Wrap(400, "私聊无法退出", errors.New("cannot leave private chat"))
	}

	// 如果是群主，解散群聊
	if conversation.OwnerID == userID {
		return db.Transaction(func(tx *gorm.DB) error {
			// 删除所有成员记录
			if err := tx.Where("conversation_id = ?", conversationID).Delete(&model.ConversationUser{}).Error; err != nil {
				return err
			}
			// 删除会话
			if err := tx.Delete(&conversation).Error; err != nil {
				return err
			}
			return nil
		})
	}

	// 普通成员直接退出
	return deleteConversationUser(db, userID, conversationID)
}
